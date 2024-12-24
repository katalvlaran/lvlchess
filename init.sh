#!/bin/bash

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Initializing Chess Bot...${NC}"

# Проверка наличия необходимых инструментов
check_requirements() {
    local missing_tools=()

    if ! command -v go &> /dev/null; then
        missing_tools+=("Go")
    fi

    if ! command -v git &> /dev/null; then
        missing_tools+=("Git")
    fi

    if [ ${#missing_tools[@]} -ne 0 ]; then
        echo -e "${RED}Error: Required tools are missing: ${missing_tools[*]}${NC}"
        exit 1
    fi
}

# Проверка версии Go
check_go_version() {
    GO_VERSION=$(go version | awk '{print $3}')
    MIN_VERSION="go1.22"
    if [[ "${GO_VERSION}" < "${MIN_VERSION}" ]]; then
        echo -e "${RED}Error: Go version must be ${MIN_VERSION} or higher${NC}"
        exit 1
    }
}

# Создание структуры проекта
create_project_structure() {
    echo -e "${YELLOW}Creating project structure...${NC}"
    
    # Основные директории
    mkdir -p cmd/bot
    mkdir -p internal/{game,ai,cache,security,monitoring,optimization,utils}
    mkdir -p configs
    mkdir -p logs
    
    # Проверка успешности создания директорий
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: Failed to create project structure${NC}"
        exit 1
    fi
}

# Инициализация Go модуля
init_go_module() {
    echo -e "${YELLOW}Initializing Go module...${NC}"
    
    go mod init github.com/katalvlaran/telega-shess
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: Failed to initialize Go module${NC}"
        exit 1
    }
    
    # Установка зависимостей
    go get github.com/notnil/chess
    go get github.com/stretchr/testify
    go get github.com/prometheus/client_golang/prometheus
    go get golang.org/x/crypto/bcrypt
    
    go mod tidy
}

# Создание конфигурационных файлов
create_config_files() {
    echo -e "${YELLOW}Creating configuration files...${NC}"
    
    # Создание .env файла
    cat > .env << EOL
# Bot Configuration
BOT_TOKEN=your_bot_token_here
DIFFICULTY=medium
MAX_GAMES=100
CACHE_SIZE=10000
ENV=development

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json

# Security Configuration
RATE_LIMIT=60
TOKEN_EXPIRATION=24h
EOL

    # Создание .gitignore
    cat > .gitignore << EOL
# Binaries and build artifacts
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out

# Environment and configuration
.env
*.local
*.local.*

# IDE specific files
.idea/
.vscode/
*.swp
*.swo

# Logs
*.log
logs/

# Dependencies
vendor/

# OS specific
.DS_Store
Thumbs.db
EOL

    # Создание README.md если его нет
    if [ ! -f README.md ]; then
        cat > README.md << EOL
# Chess Bot

Шахматный бот с поддержкой различных уровней сложности и анализом позиций.

## Установка и запуск

\`\`\`bash
./init.sh    # Инициализация проекта
go run cmd/bot/main.go  # Запуск бота
\`\`\`

## Конфигурация

Настройте параметры в файле .env перед запуском.
EOL
    fi
}

# Проверка и создание файла логов
setup_logging() {
    echo -e "${YELLOW}Setting up logging...${NC}"
    
    touch logs/chessbot.log
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: Failed to create log file${NC}"
        exit 1
    fi
}

# Запуск тестов
run_tests() {
    echo -e "${YELLOW}Running tests...${NC}"
    
    go test ./... -v
    if [ $? -ne 0 ]; then
        echo -e "${YELLOW}Warning: Some tests failed${NC}"
    }
}

# Основной процесс
main() {
    echo -e "${YELLOW}Starting initialization...${NC}"
    
    check_requirements
    check_go_version
    create_project_structure
    init_go_module
    create_config_files
    setup_logging
    run_tests
    
    echo -e "${GREEN}Initialization completed successfully!${NC}"
    echo -e "${YELLOW}Please update .env file with your bot token${NC}"
}

# Запуск скрипта
main 