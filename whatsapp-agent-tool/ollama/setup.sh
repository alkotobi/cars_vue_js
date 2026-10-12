#!/bin/bash
# Set up the local Ollama models that drive the WhatsApp agent via opencode.
# Run this AFTER your VPN / proxy is up (registry.ollama.ai must be reachable):
#   bash ollama/setup.sh
#
# Pulls the base models, then creates the -ctx64 derived models (64K context,
# which opencode requires) that opencode.json points at.
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"

for base in qwen3.6:27b gemma4:26b; do
  if ollama list | grep -q "^${base}"; then
    echo ">> ${base} already present, skipping pull"
  else
    echo ">> pulling ${base} (this is the big download)"
    ollama pull "${base}"
  fi
done

echo ">> creating qwen3.6:27b-ctx64"
ollama create qwen3.6:27b-ctx64 -f "$HERE/Modelfile.qwen3.6-27b"
echo ">> creating gemma4:26b-ctx64"
ollama create gemma4:26b-ctx64 -f "$HERE/Modelfile.gemma4-26b"

echo ">> done. Installed models:"
ollama list