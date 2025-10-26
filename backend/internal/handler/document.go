package handler

import (
	"backend/internal/service"
	"context"
	"errors"
	"io"

	"github.com/go-chi/jwtauth/v5"
)

// UploadDocument реализует StrictServerInterface.
func (h *handler) UploadDocument(ctx context.Context, request UploadDocumentRequestObject) (UploadDocumentResponseObject, error) {
	_, claims, _ := jwtauth.FromContext(ctx)
	userID := int64(claims["user_id"].(float64))

	// request.Body - это multipart.Reader, его нужно обработать
	multipartReader := request.Body
	var fileContent []byte
	var filename string

	for {
		part, err := multipartReader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return UploadDocument400Response{}, nil // Возвращаем 400, если форма невалидна
		}

		if part.FormName() == "file" {
			filename = part.FileName()
			fileContent, err = io.ReadAll(part)
			if err != nil {
				return nil, err // Внутренняя ошибка при чтении файла
			}
			part.Close()
			break // Нашли файл, выходим
		}
		part.Close()
	}

	if fileContent == nil {
		return UploadDocument400Response{}, nil // Файл не был найден в запросе
	}

	doc, err := h.service.UploadAndProcessDocument(ctx, userID, filename, fileContent)
	if err != nil {
		return nil, err
	}

	return UploadDocument201JSONResponse{
		Id:       &doc.ID,
		UserID:   &doc.UserID,
		Filename: &doc.Filename,
	}, nil
}

// ListUserDocuments реализует StrictServerInterface.
func (h *handler) ListUserDocuments(ctx context.Context, request ListUserDocumentsRequestObject) (ListUserDocumentsResponseObject, error) {
	_, claims, _ := jwtauth.FromContext(ctx)
	userID := int64(claims["user_id"].(float64))

	docs, err := h.service.ListUserDocuments(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Конвертируем []service.Document в []Document (сгенерированный тип)
	responseDocs := make(ListUserDocuments200JSONResponse, len(docs))
	for i, d := range docs {
		responseDocs[i] = Document{
			Id:       &d.ID,
			UserID:   &d.UserID,
			Filename: &d.Filename,
		}
	}

	return responseDocs, nil
}

// DeleteDocument реализует StrictServerInterface.
func (h *handler) DeleteDocument(ctx context.Context, request DeleteDocumentRequestObject) (DeleteDocumentResponseObject, error) {
	_, claims, _ := jwtauth.FromContext(ctx)
	userID := int64(claims["user_id"].(float64))

	// documentID уже распарсен и доступен в request
	docID := request.DocumentID

	if err := h.service.DeleteUserDocument(ctx, userID, docID); err != nil {
		// Можно добавить обработку service.ErrDocumentNotFound, если нужно
		return nil, err
	}

	return DeleteDocument204Response{}, nil
}

// SearchInDocument реализует StrictServerInterface.
func (h *handler) SearchInDocument(ctx context.Context, request SearchInDocumentRequestObject) (SearchInDocumentResponseObject, error) {
	_, claims, _ := jwtauth.FromContext(ctx)
	userID := int64(claims["user_id"].(float64))

	docID := request.DocumentID
	query := request.Body.Query

	results, err := h.service.SearchInDocument(ctx, userID, docID, query)
	if err != nil {
		return nil, err
	}

	responseResults := make(SearchInDocument200JSONResponse, len(results))
	for i, r := range results {
		responseResults[i] = SearchResult{
			Id:         &r.ID,
			DocumentID: &r.DocumentID,
			Text:       &r.Text,
			Distance:   &r.Distance,
		}
	}

	return responseResults, nil
}

// GetDocumentByID реализует StrictServerInterface.
func (h *handler) GetDocumentByID(ctx context.Context, request GetDocumentByIDRequestObject) (GetDocumentByIDResponseObject, error) {
	_, claims, _ := jwtauth.FromContext(ctx)
	userID := int64(claims["user_id"].(float64))

	docID := request.DocumentID

	doc, err := h.service.GetDocumentByID(ctx, userID, docID)
	if err != nil {
		if errors.Is(err, service.ErrDocumentNotFound) {
			return GetDocumentByID404Response{}, nil
		}
		return nil, err
	}

	return GetDocumentByID200JSONResponse{
		Id:       &doc.ID,
		UserID:   &doc.UserID,
		Filename: &doc.Filename,
	}, nil
}
