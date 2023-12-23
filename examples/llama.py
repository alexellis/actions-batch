#!/bin/bash

## Example by Alex Ellis

# Download a model from HuggingFace and run an inference with 300 tokens.

## Adapted from
# https://swharden.com/blog/2023-07-29-ai-chat-locally-with-python/
# Model: https://huggingface.co/TheBloke/Llama-2-7B-Chat-GGUF

python -m venv .venv
chmod +x .venv/bin/activate
.venv/bin/activate

pip install llama-cpp-python

pip install --upgrade huggingface_hub

huggingface-cli download \
  TheBloke/Llama-2-7B-Chat-GGUF \
  config.json llama-2-7b-chat.Q5_K_M.gguf --local-dir .

cat > main.py <<EOF
#!/bin/python

# load the large language model file
from llama_cpp import Llama
LLM = Llama(model_path="./llama-2-7b-chat.Q5_K_M.gguf")
   
# create a text prompt
prompt = "Q: What are the names of the days of the week? A:"

# generate a response (takes several seconds)
output = LLM(prompt,max_tokens=300, stop=[])
   
# display the response
print(output["choices"][0]["text"])
print(output["choices"])

EOF

mkdir -p uploads

chmod +x main.py
cp ./main.py > uploads/
./main.py > uploads/output.txt

