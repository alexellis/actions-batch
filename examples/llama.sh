#!/bin/bash

## Example by Alex Ellis

# Download a model from HuggingFace and run an inference against
# a list of questions in a text file, using 300 tokens.

## Adapted from
# https://swharden.com/blog/2023-07-29-ai-chat-locally-with-python/
# Model: https://huggingface.co/TheBloke/Llama-2-7B-Chat-GGUF

cat > questions.txt <<EOF
Q: What are the names of the days of the week? A:
Q: Summarise your training data in one sentence. A:
Q: What is the best way to learn? A:
Q: Who was known as the Stoic Emperor? A:
Q: What would Marcus Aurelius say was the key to peace of mind? A:
EOF

python -m venv .venv
chmod +x .venv/bin/activate
.venv/bin/activate

pip install llama-cpp-python

pip install --upgrade huggingface_hub

# This is the longest part of the job:

huggingface-cli download \
  TheBloke/Llama-2-7B-Chat-GGUF \
  config.json llama-2-7b-chat.Q5_K_M.gguf --local-dir .


cat > main.py <<EOF
#!/bin/python

import time

# load the large language model file

from llama_cpp import Llama
LLM = Llama(model_path="./llama-2-7b-chat.Q5_K_M.gguf")

questions = []
with open("questions.txt") as f:
    questions = f.readlines()

for question in questions:
    # create a text prompt
    prompt = question.strip()
    if prompt == "":
        continue

    startTime = time.time()
    # generate a response (takes several seconds)
    output = LLM(prompt,max_tokens=300, stop=[])
   
    duration = time.time() - startTime

    print("")
    print("[{}] Q: {}".format(duration, prompt))
    print(output["choices"][0]["text"])
    print("")

EOF

mkdir -p uploads

chmod +x main.py
cp ./main.py uploads/

# fd1 is stdout
# fd2 is stderr which has a lot of noise, so we're redirecting it to /dev/null
./main.py 1> uploads/output.txt 2> uploads/output-stderr.txt

# ./main.py > uploads/output.txt

# Also display the results in the job log
cat uploads/output.txt
