#!/bin/bash
set -euo pipefail

if [ "${CLAUDE_CODE_REMOTE:-}" != "true" ]; then
  exit 0
fi

pip install \
  pytest \
  flake8 \
  numpy \
  pandas \
  scikit-learn==1.3.0 \
  matplotlib \
  seaborn \
  scipy \
  joblib==1.2.0 \
  tqdm==4.62.3 \
  logzero==1.7.0 \
  requests \
  python-dotenv \
  watermark \
  xgboost \
  lightgbm \
  networkx \
  pyarrow
