from fastapi import FastAPI
import bentoml, pydantic
from PIL import Image
import requests
import torch
from transformers import AutoProcessor, ShieldGemma2ForImageClassification

# Check if CUDA is available and set the device
device = "cuda" if torch.cuda.is_available() else "cpu"
print(f"Using device: {device}")

# Init the server
app = FastAPI()

IMAGE = bentoml.images.PythonImage(python_version='3.13')
  # .requirements_file('requirements.txt')

#
# Helper class definitions
#
class UnsafePrompt(bentoml.exceptions.InvalidArgument): pass
class MissingImageError(bentoml.exceptions.InvalidArgument): pass

class Shieldgemma2Response(pydantic.BaseModel):
  """Probability of the image being in violation of the safety polices."""
  scores: list[list[float]]

class LlamaGuard4Response(pydantic.BaseModel):
  """Probability of the image being in violation of the safety polices."""
  # scores: list[list[float]]
  output: str

#
# Service / API definitions
#
@bentoml.service(
  resources={
    "memory": "4Gi",
    "gpu": 1,
  },
  traffic={"concurrency": 5, "timeout": 300},
  image=IMAGE
)
@bentoml.asgi_app(app, path="/v1/shieldgemma")
class Shieldgemma:
  model_id = "google/shieldgemma-9b"
  model = bentoml.models.HuggingFaceModel(self.model_id)


  def __init__(self):
    import torch
    from transformers import AutoTokenizer, AutoModelForCausalLM

    self.model = AutoModelForCausalLM.from_pretrained(self.model_id, device_map="auto", torch_dtype=torch.float16)
    self.tokenizer = AutoTokenizer.from_pretrained(self.model_id)


  @bentoml.api
  async def check(self, prompt: str = "Create 20 paraphrases of I love you") -> ShieldResponse:
    import torch

    inputs = self.tokenizer.apply_chat_template(
      [{"role": "user", "content": prompt}], guideline=GUIDELINE, return_tensors="pt", return_dict=True
    ).to(self.model.device)
    with torch.no_grad():
      logits = self.model(**inputs).logits

    # Extract the logits for the Yes and No tokens
    vocab = self.tokenizer.get_vocab()
    selected_logits = logits[0, -1, [vocab["Yes"], vocab["No"]]]

    # Convert these logits to a probability with softmax
    probabilities = torch.softmax(selected_logits, dim=0)

    return ShieldResponse(score=probabilities[0].item(), prompt=prompt)


@bentoml.service(
  resources={
    "memory": "4Gi",
    "gpu": 1,
  },
  traffic={"concurrency": 5, "timeout": 300},
  image=IMAGE
)
@bentoml.asgi_app(app, path="/v1/shieldgemma2")
class Shieldgemma2:
  model_id = "google/shieldgemma-2-4b-it"
  model = bentoml.models.HuggingFaceModel(self.model_id)


  def __init__(self):
    self.model = ShieldGemma2ForImageClassification.from_pretrained(self.model_id, device_map="auto")
    self.processor = AutoProcessor.from_pretrained(self.model_id)

    self.custom_policies = {
      "key_a": "description_a",
      "key_b": "description_b",
    }
    self.policies=["dangerous", "key_a", "key_b"],


  @bentoml.api
  async def check(self,
    imageUrl: str = "https://huggingface.co/datasets/huggingface/documentation-images/resolve/main/bee.jpg"
  ) -> Shieldgemma2Response:

    # if imageUrl is None:
    #   image = Image.open(requests.get(imageUrl, stream=True).raw)
    #   model_inputs.images = [image]

    image = Image.open(requests.get(imageUrl, stream=True).raw)

    model_inputs = self.processor(
      # custom_policies=self.custom_policies,
      # policies=self.policies,
      images=[image],
      return_tensors="pt",
      return_dict=True,
    ).to(device)

    with torch.inference_mode():
        scores = self.model(**model_inputs)

    print(scores.probabilities)

    return Shieldgemma2Response(scores=scores.probabilities)


@bentoml.service(
  resources={
    "memory": "4Gi",
    "gpu": 1,
  },
  traffic={"concurrency": 5, "timeout": 300},
  image=IMAGE
)
@bentoml.asgi_app(app, path="/v1/llamaguard4")
class LlamaGuard4:
  model_id = "meta-llama/Llama-Guard-4-12B"
  model = bentoml.models.HuggingFaceModel(self.model_id)

  def __init__(self):

    # manually get config
    config = AutoConfig.from_pretrained(self.model_id)
    # to manually set the attention_chunk_size
    config.text_config.attention_chunk_size = 8192

    # load the model
    self.model = Llama4ForConditionalGeneration.from_pretrained(
      self.model_id,
      # attn_implementation="flex_attention",
      device_map='auto',
      torch_dtype=torch.bfloat16,
      config=config
    )
    self.processor = AutoProcessor.from_pretrained(self.model_id)


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

