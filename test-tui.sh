#!/bin/bash
# Script para probar la TUI de Synkro
# Simula entrada de teclado

# Primero, verificar que haya datos en la base de datos
if [ ! -f "memory.db" ]; then
    echo "Iniciando base de datos..."
    go run ./cmd/synkro/ init
fi

# Agregar algunos datos de prueba
echo "Agregando memorias de prueba..."
go run ./cmd/synkro/ add --title "Prueba TUI 1" --content "Esta es una memoria de prueba para la TUI" --type note
go run ./cmd/synkro/ add --title "Prueba TUI 2" --content "Otra memoria de prueba" --type decision
go run ./cmd/synkro/ add --title "Prueba TUI 3" --content "Tercera memoria de prueba" --type task

echo "Memorias agregadas:"
go run ./cmd/synkro/ list

echo ""
echo "=== INSTRUCCIONES PARA PROBAR LA TUI ==="
echo ""
echo "Navegación:"
echo "  ↑ / k  - Mover hacia arriba"
echo "  ↓ / j  - Mover hacia abajo"
echo "  Enter  - Ver detalles/grafo"
echo "  /       - Buscar"
echo "  g       - Ver grafo"
echo "  Ctrl+C  - Salir"
echo ""
echo "Para salir de la TUI presiona Ctrl+C"
echo ""
echo "Presiona Enter para lanzar la TUI..."
read

go run ./cmd/synkro/ tui
