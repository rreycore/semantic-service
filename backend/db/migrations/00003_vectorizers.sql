-- +goose Up
-- +goose StatementBegin

select ai.create_vectorizer(
     'chunks'::regclass,
     name => 'document_chunks_vectorizer',
     loading => ai.loading_column('text'),
     embedding => ai.embedding_openai('text-embedding-3-small', 768),
     chunking => ai.chunking_none(),
     destination => ai.destination_column('embedding'),
     formatting => ai.formatting_python_template('passage: $chunk'),
     if_not_exists => true
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
select ai.drop_vectorizer('document_chunks_vectorizer');
-- +goose StatementEnd
