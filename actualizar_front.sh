#!/bin/bash
cd ~/qisur-challenge

echo "Obteniendo últimos cambios..."
git pull

echo "Reconstruyendo backend..."
docker compose up -d --build
docker compose restart qisur-gateway

echo "Construyendo frontend estático..."
cd tracker-ui
npm install
npm run build

echo "Copiando estáticos al Nginx del VPS..."
mkdir -p /var/www/qisur
rm -rf /var/www/qisur/*
cp -r dist/* /var/www/qisur/

echo "¡Qisur Challenge actualizado con éxito!"
