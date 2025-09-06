from transformers import AutoProcessor, ShieldGemma2ForImageClassification
from PIL import Image
import requests
import torch

# Check if CUDA is available and set the device
# device = "cuda" if torch.cuda.is_available() else "cpu"
device = "cpu"
print(f"Using device: {device}")

model_id = "google/shieldgemma-2-4b-it"

url = "https://huggingface.co/datasets/huggingface/documentation-images/resolve/main/bee.jpg"
image = Image.open(requests.get(url, stream=True).raw)

# Move the model to the selected device
model = ShieldGemma2ForImageClassification.from_pretrained(model_id).to(device).eval()
processor = AutoProcessor.from_pretrained(model_id)

# Process the image and move the input tensors to the selected device
model_inputs = processor(images=[image], return_tensors="pt")
model_inputs = {k: v.to(device) for k, v in model_inputs.items()}

with torch.inference_mode():
    scores = model(**model_inputs)

print(scores.probabilities)