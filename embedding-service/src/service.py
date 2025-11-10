import os

import batched
import numpy as np
import torch
from sentence_transformers import SentenceTransformer


class Service:
    def __init__(self):
        model = os.environ["MODEL_NAME"]
        dimensions = int(os.environ["EMBEDDING_DIMENSIONS"])
        if torch.backends.mps.is_available():
            device = "mps"
        else:
            device = "cpu"
        self.model = SentenceTransformer(
            model_name_or_path=model,
            device=device,
            truncate_dim=dimensions,
        )

    @batched.dynamically(batch_size=100, timeout_ms=100)
    def gen_embeddings(self, documents: str | list[str]) -> list[list[float]]:
        embeddings: np.ndarray = self.model.encode(documents)
        embeddings_32 = embeddings.astype(np.float32)
        return embeddings_32.tolist()
