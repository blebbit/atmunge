# in a file named `service.py`
import bentoml
import torch
from transformers import AutoProcessor, ShieldGemma2ForImageClassification
from PIL import Image
import requests
import numpy as np
from typing import List

# Import the same Pydantic schemas from the FastAPI example
from schema import ChatCompletionRequest, ChatCompletionResponse, ResponseMessage, Choice
import time

MODEL_ID = "google/shieldgemma-2-4b-it"

@bentoml.service(
    resources={"gpu": 1, "gpu_type": "nvidia-l4"}, # Recommended GPU [26]
    traffic={"timeout": 30},
)
class ShieldGemmaService:
    def __init__(self) -> None:
        self.device = "cuda" if torch.cuda.is_available() else "cpu"
        self.model = ShieldGemma2ForImageClassification.from_pretrained(MODEL_ID).to(self.device).eval()
        self.processor = AutoProcessor.from_pretrained(MODEL_ID)
        print(f"Model '{MODEL_ID}' loaded on {self.device}")

    @bentoml.api
    def create_chat_completion(self, request: ChatCompletionRequest) -> ChatCompletionResponse:
        user_message = request.messages[-1]
        image_url = None
        for part in user_message.content:
            if part.type == "image_url" and part.image_url:
                image_url = part.image_url.url
                break

        if not image_url:
            # In a real service, you'd use BentoML's exception handling
            raise bentoml.exceptions.BentoMLException("No image_url found.")

        try:
            raw_image = Image.open(requests.get(image_url, stream=True).raw)
            model_inputs = self.processor(images=[raw_image], return_tensors="pt").to(self.device)

            with torch.inference_mode():
                scores = self.model(**model_inputs)

            probabilities = scores.probabilities.cpu().numpy()

            # --- Adapter Logic (same as before) ---
            violation_prob = probabilities
            is_violation = violation_prob > 0.5
            result_text = "Yes" if is_violation else "No"
            confidence = violation_prob if is_violation else (1 - violation_prob)
            response_content = f'{{"result": "{result_text}", "confidence": {confidence:.4f}, "policy": "default"}}'

            # --- Construct OpenAI-compatible Response ---
            response_message = ResponseMessage(role="assistant", content=response_content)
            choice = Choice(message=response_message)
            return ChatCompletionResponse(
                created=int(time.time()),
                model=MODEL_ID,
                choices=[choice]
            )
        except Exception as e:
            raise bentoml.exceptions.BentoMLException(f"Inference failed: {e}")