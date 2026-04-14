#!/bin/bash
# Script para crear paquete de distribución Synkro v2

VERSION="2.0.0"
PACKAGE_NAME="synkro-${VERSION}"

echo "📦 Creando paquete Synkro v2..."

# Crear directorio de paquete
mkdir -p "$PACKAGE_NAME"

# Copiar binario
cp synkro "$PACKAGE_NAME/"
chmod +x "$PACKAGE_NAME/synkro"

# Copiar documentación esencial
cp README.md "$PACKAGE_NAME/"
cp QUICKSTART.md "$PACKAGE_NAME/"
cp AGENTS.md "$PACKAGE_NAME/"
cp INSTALL.md "$PACKAGE_NAME/"
cp EMBEDDINGS.md "$PACKAGE_NAME/"
cp TUI.md "$PACKAGE_NAME/"

# Crear script de instalación
cat > "$PACKAGE_NAME/install.sh" << 'INSTALL_EOF'
#!/bin/bash

echo "🚀 Instalando Synkro v2..."

# Detectar plataforma
OS=$(uname -s)
ARCH=$(uname -m)

# Crear directorio de instalación
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    echo "⚠️  Necesito permisos para instalar en $INSTALL_DIR"
    sudo mkdir -p "$INSTALL_DIR" 2>/dev/null || INSTALL_DIR="$HOME/.local/bin"
    export PATH="$PATH:$INSTALL_DIR"
fi

# Copiar binario
if [ -w "$INSTALL_DIR" ]; then
    cp synkro "$INSTALL_DIR/"
else
    sudo cp synkro "$INSTALL_DIR/"
fi

chmod +x "$INSTALL_DIR/synkro"

# Inicializar base de datos
echo ""
echo "📊 Inicializando base de datos..."
"$INSTALL_DIR/synkro" init

echo ""
echo "✅ Synkro v2 instalado exitosamente!"
echo ""
echo "📚 Para empezar:"
echo "  synkro --help"
echo "  synkro init"
echo "  synkro tui"
echo ""
echo "📖 Documentación:"
echo "  README.md       - Documentación completa"
echo "  QUICKSTART.md   - Guía rápida"
echo "  AGENTS.md      - Guía para agentes de IA"
echo "  INSTALL.md      - Instalación MCP"
echo ""
INSTALL_EOF

chmod +x "$PACKAGE_NAME/install.sh"

# Crear README específico del paquete
cat > "$PACKAGE_NAME/README.txt" << 'PKG_EOF'
Synkro v${VERSION} - Motor de Contexto Inteligente para LLMs
====================================================================

🚀 INSTALACIÓN ONE-SHOT

  ./install.sh

📋 LO QUE INCLUYE ESTE PAQUETE

  ✅ synkro           - Binario completo (8.7MB)
  ✅ MCP server       - Integrado en binario
  ✅ TUI profesional  - Bubble Tea + Lipgloss
  ✅ Embeddings       - TF-IDF + MiniLM support
  ✅ Documentación    - Guías completas

📚 GUÍAS DISPONIBLES

  README.md       - Documentación completa del proyecto
  QUICKSTART.md   - Guía rápida de inicio
  AGENTS.md      - Guía para integrar con agentes de IA
  INSTALL.md      - Guía de instalación MCP
  EMBEDDINGS.md   - Modelos de embeddings disponibles
  TUI.md         - Guía de la TUI profesional

🖥️  COMANDOS DISPONIBLES

  synkro init           - Inicializar base de datos
  synkro add            - Agregar memoria
  synkro list           - Listar memorias
  synkro search         - Buscar memorias
  synkro tui            - Lanzar TUI profesional
  synkro mcp            - Iniciar servidor MCP

🎯 QUICK START

  1. Instalar: ./install.sh
  2. Inicializar: synkro init
  3. Agregar memoria: synkro add --title "Test" --content "Test" --type note
  4. Lanzar TUI: synkro tui
  5. Buscar: synkro search "Test"

🔗 MÁS INFORMACIÓN

  GitHub: https://github.com/rodascaar/synkro
  Docs: https://github.com/rodascaar/synkro/tree/main#documentation

====================================================================
Synkro v${VERSION} - Build: $(date +%Y%m%d)
Estado: 100% Completado ✅
PKG_EOF

# Crear tar.gz
echo ""
echo "📦 Creando paquete tar.gz..."
tar -czf "${PACKAGE_NAME}-$(uname -s)-$(uname -m).tar.gz" "$PACKAGE_NAME"

# Crear lista de archivos
echo "📋 Lista de archivos:"
ls -lh "$PACKAGE_NAME/"

echo ""
echo "✅ Paquete creado: ${PACKAGE_NAME}-$(uname -s)-$(uname -m).tar.gz"
echo ""
echo "📦 Tamaño del paquete:"
ls -lh "${PACKAGE_NAME}-$(uname -s)-$(uname -m).tar.gz"

echo ""
echo "📋 Para instalar en otro sistema:"
echo "  1. Copiar ${PACKAGE_NAME}-$(uname -s)-$(uname -m).tar.gz"
echo "  2. Extraer: tar -xzf ${PACKAGE_NAME}-$(uname -s)-$(uname -m).tar.gz"
echo "  3. Instalar: cd $PACKAGE_NAME && ./install.sh"
echo ""
