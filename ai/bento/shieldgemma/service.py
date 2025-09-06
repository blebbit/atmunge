import bentoml, pydantic
import torch
from transformers import AutoTokenizer, AutoModelForCausalLM
# from openai import AsyncOpenAI

MODEL_ID = "google/shieldgemma-2b"
IMAGE = bentoml.images.PythonImage(python_version='3.13').requirements_file('requirements.txt')

GUIDELINE = '"No Harassment": The prompt shall not contain or seek generation of content that is malicious, intimidating, bullying, or abusive content targeting another individual (e.g., physical threats, denial of tragic events, disparaging victims of violence).'
PROMPT = """You are a policy expert trying to help determine whether a user prompt is in violation of the defined safety policies.

<start_of_turn>
Human Question: {user_prompt}
<end_of_turn>

Our safety principle is defined in the below:

{safety_policy}

Does the human question violate the above principle? Your answer must start
with 'Yes' or 'No'. And then walk through step by step to be sure we answer
correctly.
"""


class ShieldgemmaResponse(pydantic.BaseModel):
  score: float
  """Probability of the prompt being in violation of the safety policy."""
  prompt: str


@bentoml.service(
  resources={
    "memory": "4Gi",
    "gpu": 1,
    # "gpu_type": "nvidia-tesla-t4"
  },
  traffic={"concurrency": 5, "timeout": 300},
  # envs=[{'name': 'HF_TOKEN'}],
  image=IMAGE
)
class Shieldgemma:
  model = bentoml.models.HuggingFaceModel(MODEL_ID)

  def __init__(self):

    self.model = AutoModelForCausalLM.from_pretrained(self.model, device_map="auto", torch_dtype=torch.float16)
    self.tokenizer = AutoTokenizer.from_pretrained(MODEL_ID)

  @bentoml.api
  async def check(self, prompt: str = "Create 20 paraphrases of I love you") -> ShieldgemmaResponse:

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

    return ShieldgemmaResponse(score=probabilities[0].item(), prompt=prompt)
