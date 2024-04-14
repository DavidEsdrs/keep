#!/bin/bash

# Verifique se o arquivo existe
if [ -f "keeps.kps" ]; then
    # Remova o arquivo
    rm "keeps.kps"
    echo "arquivo keeps removido"
fi

# Verifique se o arquivo existe
if [ -f "info.kpsinfo" ]; then
    # Remova o arquivo
    rm "info.kpsinfo"
    echo "arquivo keeps info removido"
fi