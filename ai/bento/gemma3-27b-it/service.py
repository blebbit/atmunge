import bentoml, pydantic
from PIL import Image
import requests
import torch
from transformers import AutoProcessor, AutoModelForImageTextToText

# Check if CUDA is available and set the device
device = "cuda" if torch.cuda.is_available() else "cpu"
print(f"Using device: {device}")



# MODEL_ID = "google/gemma-3-4b-it"
# MODEL_ID = "google/gemma-3-27b-it"
MODEL_ID = "unsloth/gemma-3-27b-it-bnb-4bit"
# MODEL_ID = "MISHANM/google-gemma-3-27b-it-fp8"
IMAGE = bentoml.images.PythonImage(python_version='3.13')
  # .requirements_file('requirements.txt')

class Gemma3Response(pydantic.BaseModel):
  message: str

@bentoml.service(
  resources={
    # "cpu": 1,
    # "memory": "4Gi",
    "gpu": 1,
  },
  traffic={"concurrency": 5, "timeout": 300},
  # envs=[{'name': 'HF_TOKEN'}],
  image=IMAGE
)
class Gemma:
  model = bentoml.models.HuggingFaceModel(MODEL_ID)

  def __init__(self):

    self.model = AutoModelForImageTextToText.from_pretrained(MODEL_ID, device_map="auto")
    self.processor = AutoProcessor.from_pretrained(MODEL_ID)


  @bentoml.api
  async def chat(self,
    # messages
  ) -> Gemma3Response:

    messages = [
      {
        "role": "system",
        "content": [{"type": "text", "text": "You are a helpful assistant."}]
      },
      {
        "role": "user",
        "content": [
          {"type": "image", "url": "https://huggingface.co/datasets/huggingface/documentation-images/resolve/main/p-blog/candy.JPG"},
          {"type": "text", "text": "What animal is on the candy?"}
        ]
      }
    ]

    print("inputs:", messages)


    inputs = self.processor.apply_chat_template(
      messages,
      add_generation_prompt=True,
      tokenize=True,
      return_dict=True,
      return_tensors="pt",
    ).to(self.model.device)

    outputs = self.model.generate(**inputs, max_new_tokens=40)
    print("outputs:", outputs)
    msg = self.processor.decode(outputs[0][inputs["input_ids"].shape[-1]:])
    print(msg)

    return Gemma3Response(message=msg)