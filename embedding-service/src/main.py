from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException

from .models import (
    EmbeddingData,
    EmbeddingRequest,
    EmbeddingResponse,
    UsageData,
)
from .service import Service


@asynccontextmanager
async def lifespan(app: FastAPI):
    app.state.service = Service()
    yield


app = FastAPI(lifespan=lifespan)


@app.get("/ping")
async def ping():
    return "pong"


@app.post("/embeddings")
async def stt(req: EmbeddingRequest) -> EmbeddingResponse:
    try:
        raw_embeddings = await app.state.service.gen_embeddings.acall(req.input)
        if isinstance(req.input, str):
            embeddings_list = [raw_embeddings]
        else:
            embeddings_list = raw_embeddings
        data = []
        for idx, e in enumerate(embeddings_list):
            data.append(EmbeddingData(embedding=e, index=idx))
        return EmbeddingResponse(
            data=data, model=req.model, usage=UsageData(prompt_tokens=0, total_tokens=0)
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Internal server error: {e}")
