package handler

import (
	"backend/internal/service"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/go-chi/jwtauth/v5"
)

func (h *handler) UploadDocument(ctx context.Context, request UploadDocumentRequestObject) (UploadDocumentResponseObject, error) {
	_, claims, _ := jwtauth.FromContext(ctx)
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		h.log.Warn().Msg("JWT: user_id missing or invalid type")
		return UploadDocument400Response{}, fmt.Errorf("invalid user ID in token")
	}
	userID := int64(userIDFloat)

	multipartReader := request.Body
	if multipartReader == nil {
		h.log.Warn().Int64("user_id", userID).Msg("multipart body is nil")
		return UploadDocument400Response{}, fmt.Errorf("invalid multipart request")
	}

	var fileContent []byte
	var filename string
	foundFile := false

	for {
		part, err := multipartReader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			h.log.Error().Err(err).Int64("user_id", userID).Msg("Ошибка чтения multipart части")
			return UploadDocument400Response{}, fmt.Errorf("invalid multipart data")
		}

		partName := part.FormName()
		partFilename := part.FileName()

		if partName == "file" {
			if partFilename == "" {
				err = part.Close()
				if err != nil {
					return nil, err
				}
				h.log.Warn().Int64("user_id", userID).Msg("Пустое имя файла в части 'file'")
				return UploadDocument400Response{}, fmt.Errorf("file name is missing")
			}

			if !strings.HasSuffix(strings.ToLower(partFilename), ".txt") {
				err = part.Close()

				if err != nil {
					return nil, err
				}
				h.log.Warn().
					Int64("user_id", userID).
					Str("filename", partFilename).
					Msg("Запрещённый тип файла — разрешены только .txt")
				return UploadDocument400Response{}, fmt.Errorf("only .txt files allowed")
			}

			fileContent, err = io.ReadAll(part)
			if err != nil {
				err = part.Close()
				if err != nil {
					return nil, err
				}
				h.log.Error().Err(err).Int64("user_id", userID).Msg("Ошибка чтения содержимого файла")
				return nil, fmt.Errorf("failed to read file content")
			}

			filename = partFilename
			foundFile = true
			h.log.Info().
				Int64("user_id", userID).
				Str("filename", filename).
				Int("size_bytes", len(fileContent)).
				Msg("Файл успешно прочитан")

			err = part.Close()
			if err != nil {
				return nil, err
			}

			break
		}

		err = part.Close()
		if err != nil {
			return nil, err
		}
	}

	if !foundFile || fileContent == nil {
		h.log.Warn().Int64("user_id", userID).Msg("Файл не был найден в запросе")
		return UploadDocument400Response{}, fmt.Errorf("no file uploaded")
	}

	const maxSize = 10 * 1024 * 1024 // 10 MB
	if len(fileContent) > maxSize {
		h.log.Warn().
			Int64("user_id", userID).
			Str("filename", filename).
			Int("size_bytes", len(fileContent)).
			Msg("Файл превышает допустимый размер (10MB)")
		return UploadDocument400Response{}, fmt.Errorf("file too large: max 10MB")
	}

	h.log.Info().
		Int64("user_id", userID).
		Str("filename", filename).
		Int("size_bytes", len(fileContent)).
		Msg("Передача файла в сервис для обработки")

	doc, err := h.service.UploadDocument(ctx, userID, filename, fileContent)
	if err != nil {
		h.log.Error().
			Err(err).
			Int64("user_id", userID).
			Str("filename", filename).
			Msg("Ошибка в сервисе при обработке документа")
		return nil, err
	}

	h.log.Info().
		Int64("user_id", userID).
		Int64("document_id", doc.ID).
		Str("filename", doc.Filename).
		Msg("Документ успешно загружен и обработан")

	return UploadDocument201JSONResponse{
		Id:              doc.ID,
		UserID:          doc.UserID,
		Filename:        doc.Filename,
		NullEmbeddings:  doc.NullEmbeddings,
		TotalEmbeddings: doc.TotalEmbeddings,
	}, nil
}

func (h *handler) ListUserDocuments(ctx context.Context, request ListUserDocumentsRequestObject) (ListUserDocumentsResponseObject, error) {
	_, claims, _ := jwtauth.FromContext(ctx)
	userID := int64(claims["user_id"].(float64))

	docs, err := h.service.ListUserDocuments(ctx, userID)
	if err != nil {
		h.log.Error().Err(err).Int64("userID", userID).Msg("HANDLER ERROR: Ошибка сервиса")
		return nil, err
	}

	responseDocs := make(ListUserDocuments200JSONResponse, len(docs))
	for i, d := range docs {
		responseDocs[i] = Document{
			Id:              d.ID,
			UserID:          d.UserID,
			Filename:        d.Filename,
			NullEmbeddings:  d.NullEmbeddings,
			TotalEmbeddings: d.TotalEmbeddings,
		}
	}

	return responseDocs, nil
}

func (h *handler) DeleteDocument(ctx context.Context, request DeleteDocumentRequestObject) (DeleteDocumentResponseObject, error) {
	_, claims, _ := jwtauth.FromContext(ctx)
	userID := int64(claims["user_id"].(float64))

	docID := request.DocumentID

	if err := h.service.DeleteUserDocument(ctx, userID, docID); err != nil {
		return nil, err
	}

	return DeleteDocument204Response{}, nil
}

func (h *handler) Search(ctx context.Context, request SearchRequestObject) (SearchResponseObject, error) {
	_, claims, _ := jwtauth.FromContext(ctx)
	userID := int64(claims["user_id"].(float64))

	query := request.Body.Query

	results, err := h.service.Search(ctx, userID, query)
	if err != nil {
		return nil, err
	}

	responseResults := make(Search200JSONResponse, len(results))
	for i, r := range results {
		responseResults[i] = SearchResult{
			Id:         &r.ID,
			DocumentID: &r.DocumentID,
			Text:       &r.Text,
			Title:      &r.Title,
			Distance:   &r.Distance,
		}
	}

	return responseResults, nil

}

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

func (h *handler) GetDocumentByID(ctx context.Context, request GetDocumentByIDRequestObject) (GetDocumentByIDResponseObject, error) {
	_, claims, _ := jwtauth.FromContext(ctx)
	userID := int64(claims["user_id"].(float64))

	docID := request.DocumentID

	d, err := h.service.GetDocumentByID(ctx, userID, docID)
	if err != nil {
		if errors.Is(err, service.ErrDocumentNotFound) {
			return GetDocumentByID404Response{}, nil
		}
		return nil, err
	}

	return GetDocumentByID200JSONResponse{
		Id:              d.ID,
		UserID:          d.UserID,
		Filename:        d.Filename,
		NullEmbeddings:  d.NullEmbeddings,
		TotalEmbeddings: d.TotalEmbeddings,
	}, nil
}
