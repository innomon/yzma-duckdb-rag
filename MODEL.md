# Embedding Model Setup

This project uses **embeddinggemma** as the default embedding model.

## Download with Ollama

```bash
# Pull the embedding model
ollama pull embeddinggemma:latest

# Verify installation
ollama list | grep embeddinggemma
```

## Locate the GGUF File

After pulling, find the model file:

```bash
# Default Ollama model location
ls ~/.ollama/models/blobs/

# Or find by name
find ~/.ollama -name "*.gguf" 2>/dev/null
```

## Configure YDRAG

Set the model path in `config.yaml`:

```yaml
model: "/home/<user>/.ollama/models/blobs/<sha256-hash>"
```

Or via environment variable:

```bash
export YDRAG_MODEL="/path/to/embeddinggemma.gguf"
```

## Model Info

- **Model**: embeddinggemma:latest
- **Type**: Text embedding model
- **Format**: GGUF (for llama.cpp)
- **Dimensions**: 768 (Gemma embedding size)

## [AI Studio suggestions](https://gemini.google.com/app/a6c6afebafa39e7a)

To get the GGUF version of **EmbeddingGemma**, you can either download pre-converted files from the community or convert the official weights yourself using `llama.cpp`.

As of early 2026, the "latest" version typically refers to the **EmbeddingGemma-300M** model (based on Gemma 3). Because this model is highly optimized for on-device use, it is widely available in quantized GGUF formats.

### 1. Download Pre-Converted GGUF Files

The easiest way is to pull them from Hugging Face. Several reputable organizations and community members provide these:

* **Official/Semi-Official:** [ggml-org/embeddinggemma-300M-GGUF](https://huggingface.co/ggml-org/embeddinggemma-300M-GGUF)
* **Unsloth (Optimized):** [unsloth/embeddinggemma-300m-GGUF](https://huggingface.co/unsloth/embeddinggemma-300m-GGUF)
* **LM Studio Community:** [lmstudio-community/embeddinggemma-300m-qat-GGUF](https://huggingface.co/lmstudio-community/embeddinggemma-300m-qat-GGUF) (includes Quantization-Aware Training versions).

**To download via terminal:**

```bash
# Example using huggingface-cli
huggingface-cli download ggml-org/embeddinggemma-300M-GGUF embeddinggemma-300m-q4_0.gguf --local-dir . --local-dir-use-symlinks False

```

---

### 2. Get it via Ollama

If you prefer a managed environment, Ollama supports EmbeddingGemma natively.

```bash
ollama pull embeddinggemma

```

*Note: This downloads the model into Ollama's internal storage format, which uses GGUF under the hood.*

---

### 3. Convert Your Own (Custom Quantization)

If you need a specific quantization level (like `Q8_0` or `F16`) that isn't online, you can convert the original Hugging Face weights using `llama.cpp`.

1. **Clone llama.cpp and install requirements:**
```bash
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp
pip install -r requirements.txt

```


2. **Download the original weights:**
Download from [google/embeddinggemma-300m](https://huggingface.co/google/embeddinggemma-300m).
3. **Run the conversion script:**
```bash
python convert-hf-to-gguf.py path/to/embeddinggemma-300m \
  --outfile embeddinggemma-latest.gguf \
  --outtype q8_0

```



---

### Quick Reference: File Sizes

| Quantization | File Size | Recommended For |
| --- | --- | --- |
| **Q4_0** | ~230 MB | Best for mobile/Edge TPUs |
| **Q8_0** | ~330 MB | Best balance of speed/accuracy |
| **F16/BF16** | ~612 MB | Maximum precision for RAG |

> [!TIP]
> **Pro-Tip:** If you are using this for RAG, remember to prepend the task-specific instruction to your input text (e.g., `search_query: ` or `classification: `), as this model is instruction-tuned for different embedding tasks!

In the **EmbeddingGemma** ecosystem (based on Gemma 3), these "modes" are actually instruction prefixes that tell the model how to weight the semantic features of the text. Because it is a unified model, it needs to know if you are looking for a *needle in a haystack* (Retrieval) or *grouping like-with-like* (Clustering).

Here is the breakdown of the primary modes and their differences:

### 1. The Retrieval Pair (Search)

This is the most common use case, specifically designed for **RAG (Retrieval-Augmented Generation)**. It uses an asymmetric approach because queries and documents usually look very different.

* **`search_query:`** (or `Retrieval-query`)
* **Purpose:** Optimized for short, often incomplete strings or questions.
* **Behavior:** It focuses on the intent and what a potential "answer" might look like.


* **`search_document:`** (or `Retrieval-document`)
* **Purpose:** Used for the chunks of text in your database (knowledge base).
* **Behavior:** It focuses on the topical density and factual content.
* **Formatting:** You can often include a title for better results: `title: {title} | text: {content}`.



### 2. The Classification & Clustering Pair

These are "symmetric" tasks where you are comparing items of similar length and structure.

* **`classification:`**
* **Purpose:** Optimized for determining if a text belongs to a specific label (sentiment, spam, topic).
* **Behavior:** It highlights features that distinguish categorical differences rather than just semantic overlap.


* **`clustering:`**
* **Purpose:** Used when you want to group large sets of documents into "buckets" without predefined labels.
* **Behavior:** It creates a more "spread out" embedding space to help algorithms like K-Means find clear boundaries.



---

### Are there more modes?

Yes. Depending on your implementation (like the `sentence-transformers` library), EmbeddingGemma supports several task-specific instructions:

| Mode / Instruction | Best Use Case |
| --- | --- |
| **`question-answering:`** | When the query is specifically a direct question (e.g., "How old is the moon?"). |
| **`fact-checking:`** | Used for claim verification against a corpus of evidence. |
| **`sentence-similarity:`** (or `STS`) | Classic semantic similarity. Use this when you want to know if Sentence A means the same as Sentence B. |
| **`code-retrieval:`** | Specifically tuned for finding code snippets based on natural language descriptions. |

### Why does this matter?

If you use the `search_query:` prefix to embed your entire database (documents), your search accuracy will actually **drop**. The model will be looking for "question-like" features in your documents where none exist.

**Rule of Thumb:**

1. **For your Database:** Use `search_document:`.
2. **For the User's Query:** Use `search_query:`.
3. **For everything else:** Use the mode that matches your task name exactly.

