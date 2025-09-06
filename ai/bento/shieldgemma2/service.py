import bentoml, pydantic
from PIL import Image
import requests
import torch
from transformers import AutoProcessor, ShieldGemma2ForImageClassification

# Check if CUDA is available and set the device
device = "cuda" if torch.cuda.is_available() else "cpu"
print(f"Using device: {device}")

MODEL_ID = "google/shieldgemma-2-4b-it"
IMAGE = bentoml.images.PythonImage(python_version='3.13')
  # .requirements_file('requirements.txt')

class Shieldgemma2Response(pydantic.BaseModel):
  """Probability of the image being in violation of the safety polices."""
  scores: list[list[float]]

@bentoml.service(
  resources={
    # "cpu": 1,
    # "memory": "4Gi",
    "gpu": 1,
    # "gpu_type": "nvidia-l4"
  },
  traffic={"concurrency": 5, "timeout": 300},
  # envs=[{'name': 'HF_TOKEN'}],
  image=IMAGE
)
class Gemma:
  model = bentoml.models.HuggingFaceModel(MODEL_ID)

  def __init__(self):


    self.model = ShieldGemma2ForImageClassification.from_pretrained(MODEL_ID, device_map="auto")
    self.processor = AutoProcessor.from_pretrained(MODEL_ID)

    self.custom_policies = {
      "key_a": "description_a",
      "key_b": "description_b",
    }
    self.policies=["dangerous", "key_a", "key_b"],


  @bentoml.api
  async def check(self,
    imageUrl: str = "https://huggingface.co/datasets/huggingface/documentation-images/resolve/main/bee.jpg"
    # customPolicies: dict(str) = {}
    # policies: list(str) = []
  ) -> Shieldgemma2Response:

    image = Image.open(requests.get(imageUrl, stream=True).raw)

    model_inputs = self.processor(
      # custom_policies=self.custom_policies,
      # policies=self.policies,
      images=[image],
      return_tensors="pt",
      return_dict=True,
    ).to(device)

    # if imageUrl is not None:
    #   image = Image.open(requests.get(imageUrl, stream=True).raw)
    #   model_inputs.images = [image]

    with torch.inference_mode():
        scores = self.model(**model_inputs)

    print("scores:", scores)
    print("predictions:", scores.probabilities)

    return Shieldgemma2Response(scores=scores.probabilities)
