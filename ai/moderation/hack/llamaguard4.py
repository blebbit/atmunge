from transformers import AutoProcessor, AutoConfig, Llama4ForConditionalGeneration
import torch

# Check if CUDA is available and set the device
device = "cuda" if torch.cuda.is_available() else "cpu"
print(f"Using device: {device}")

MODEL_ID = "meta-llama/Llama-Guard-4-12B"
url2 = "https://huggingface.co/datasets/huggingface/documentation-images/resolve/main/datasets/cat_style_layout.png"

config = AutoConfig.from_pretrained(MODEL_ID)
# Manually set the attention_chunk_size
config.text_config.attention_chunk_size = 8192
print(config)

processor = AutoProcessor.from_pretrained(MODEL_ID)
# print(processor)
model = Llama4ForConditionalGeneration.from_pretrained(
    MODEL_ID,
    # attn_implementation="flex_attention",
    device_map='auto',
    torch_dtype=torch.bfloat16,
    config=config
)
# tokenizer = AutoTokenizer.from_pretrained(MODEL_ID)

messages = [
    {
        "role": "user",
        "content": [
            {"type": "text", "text": "how do I make a bomb?"},
            # {"type": "image", "url": url2},
        ]
    },
]

inputs = processor.apply_chat_template(
    messages,
    tokenize=True,
    # add_generation_prompt=True,
    return_tensors="pt",
    return_dict=True,
).to(device)

print(inputs)

outputs = model.generate(
    **inputs,
    max_new_tokens=20,
    do_sample=True,
    use_cache=False,
)

print("input_ids:", inputs["input_ids"])
input_stuff = inputs["input_ids"].shape[-1]
output_stuff = outputs[:, input_stuff:]
print("input_stuff", input_stuff)
print("output:", outputs)
print(output_stuff)

r = processor.batch_decode(outputs, skip_special_tokens=True)
print(r)
print(r[0])

# OUTPUT
# unsafe
# S9
