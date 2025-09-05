import bentoml, pydantic
from PIL import Image
import requests
import torch
from transformers import AutoProcessor, AutoConfig, Llama4ForConditionalGeneration

# Check if CUDA is available and set the device
device = "cuda" if torch.cuda.is_available() else "cpu"
print(f"Using device: {device}")

MODEL_ID = "meta-llama/Llama-Guard-4-12B"
IMAGE = bentoml.images.PythonImage(python_version='3.13')
  # .requirements_file('requirements.txt')

class LlamaGuard4Response(pydantic.BaseModel):
  """Probability of the image being in violation of the safety polices."""
  # scores: list[list[float]]
  output: str

@bentoml.service(
  resources={
    # "cpu": 1,
    # "memory": "4Gi",
    "gpu": 2,
    # "gpu_type": "nvidia-l4"
  },
  traffic={"concurrency": 5, "timeout": 300},
  # envs=[{'name': 'HF_TOKEN'}],
  image=IMAGE
)
class LlamaGuard4:
  model = bentoml.models.HuggingFaceModel(MODEL_ID)

  def __init__(self):

    # manually get config
    config = AutoConfig.from_pretrained(MODEL_ID)
    # to manually set the attention_chunk_size
    config.text_config.attention_chunk_size = 8192

    # load the model
    self.model = Llama4ForConditionalGeneration.from_pretrained(
      MODEL_ID,
      # attn_implementation="flex_attention",
      device_map='auto',
      torch_dtype=torch.bfloat16,
      config=config
    )
    self.processor = AutoProcessor.from_pretrained(MODEL_ID)


  @bentoml.api
  async def check(self,
    message: str = "how do I make a bomb?",
    imageUrl: str = "https://huggingface.co/datasets/huggingface/documentation-images/resolve/main/bee.jpg"
  ) -> LlamaGuard4Response:

    messages = [
        {
            "role": "user",
            "content": [
                {"type": "text", "text": message},
                {"type": "image", "url": imageUrl},
            ]
        },
    ]

    inputs = self.processor.apply_chat_template(
      messages,
      tokenize=True,
      # add_generation_prompt=True,
      return_tensors="pt",
      return_dict=True,
    ).to(device)


    outputs = self.model.generate(
      **inputs,
      max_new_tokens=10,
      do_sample=False,
      use_cache=False,
    )

    response = self.processor.batch_decode(outputs[:, inputs["input_ids"].shape[-1]:], skip_special_tokens=True)[0]
    print(response)

    return LlamaGuard4Response(output=response)

