import bentoml
from bentoml.models import HuggingFaceModel
from transformers import AutoModelForSequenceClassification, AutoTokenizer

@bentoml.service(resources={"cpu": "200m", "memory": "512Mi"})
class MyService:
    # Specify a model from HF with its ID
    # model_path = HuggingFaceModel("google/shieldgemma-2-4b-it")
    model_path = HuggingFaceModel("meta-llama/Llama-Guard-4-12B")

    def __init__(self):
        # Load the actual model and tokenizer within the instance context
        self.model = AutoModelForSequenceClassification.from_pretrained(self.model_path)
        self.tokenizer = AutoTokenizer.from_pretrained(self.model_path)

svc = MyService()

