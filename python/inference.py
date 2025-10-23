from transformers import AutoModelForCausalLM, AutoTokenizer
import torch

model_name = "Qwen/Qwen2.5-7B-Instruct"  # 或本地路径
tokenizer = AutoTokenizer.from_pretrained(model_name)
model = AutoModelForCausalLM.from_pretrained(
    model_name,
    torch_dtype=torch.float16,  # 使用半精度节省显存
    device_map="auto"  # 自动分配到 GPU
)

messages = [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "请解释量子计算的基本原理。"}
]
text = tokenizer.apply_chat_template(messages, tokenize=False, add_generation_prompt=True)
model_inputs = tokenizer([text], return_tensors="pt").to(model.device)
generated_ids = model.generate(
    model_inputs.input_ids,
    max_new_tokens=512,
    temperature=0.7,
    do_sample=True
)
generated_ids = generated_ids[:, model_inputs.input_ids.shape[-1]:]
response = tokenizer.batch_decode(generated_ids, skip_special_tokens=True)[0]
print(response)