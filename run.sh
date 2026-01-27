#!/bin/bash
export CGO_ENABLED=1
# 定义变量
GO_CMD="go"
DEFAULT_CONFIG_FILE="/mnt/d/coder/Go/src/ssh-proxy-plus/runtime/conf.yaml"
RACE_FLAG="-race"
DEBUG_FLAG="-debug"
ENTRY_POINT="cmd/main.go"
OUTPUT_BINARY="./runtime/build/proxy-plus.run"


# 显示帮助信息
show_help() {
    echo "Usage: $0 <command> [options]"
    echo "Commands:"
    echo "  run     - Run the application (default)"
    echo "  build   - Build and run the application"
    echo "Options:"
    echo "  -c <config> - Specify config file path (default: $DEFAULT_CONFIG_FILE)"
    echo "  -d         - Enable debug mode (program flag)"
    echo "  -r         - Enable race detector (build flag)"
    echo "  -dev       - Enable dev mode (adds -tags=dev)"
    echo "Examples:"
    echo "  $0 run -c ./config.yaml -d -dev"
    echo "  $0 build -r -d -dev"
    exit 1
}

# 初始化变量
MODE="build"
CONFIG_FILE="$DEFAULT_CONFIG_FILE"
DEBUG=false
RACE=false
DEV_MODE=false
RUN_AFTER_BUILD=true

# 解析参数
while [[ $# -gt 0 ]]; do
    case "$1" in
        run|build)
            MODE="$1"
            shift
            ;;
        -c)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -d)
            DEBUG=true
            shift
            ;;
        -r)
            RACE=true
            shift
            ;;
        -dev)
            DEV_MODE=true
            shift
            ;;
        -h|--help)
            show_help
            ;;
        *)
            echo "❌ Error: Unknown option $1" >&2
            show_help
            ;;
    esac
done

# 判断是否为相对路径（不以 / 开头）
if [[ "$CONFIG_FILE" != /* ]]; then
    CONFIG_FILE="$(pwd)/$CONFIG_FILE"
fi


# 构建编译参数
BUILD_ARGS=()
if [ "$RACE" = true ]; then
    BUILD_ARGS+=("$RACE_FLAG")
fi
if [ "$DEV_MODE" = true ]; then
    BUILD_ARGS+=("-tags=dev")
fi

# 构建运行参数
RUN_ARGS=()
if [ "$DEBUG" = true ]; then
    RUN_ARGS+=("$DEBUG_FLAG")
fi

# 根据模式执行不同操作
case "$MODE" in
    run)
        echo "🚀 Running application..."
        RUN_CMD=("$GO_CMD" "run" "${BUILD_ARGS[@]}" "$ENTRY_POINT" "${RUN_ARGS[@]}" -c "$CONFIG_FILE")
        # shellcheck disable=SC2145
        echo "Command: ${RUN_CMD[@]}"
        "${RUN_CMD[@]}"
        ;;
    build)
        echo "🔨 Building binary..."
        BUILD_CMD=("$GO_CMD" "build" "${BUILD_ARGS[@]}" -o "$OUTPUT_BINARY" "$ENTRY_POINT")
        # shellcheck disable=SC2145
        echo "Command: ${BUILD_CMD[@]}"
        "${BUILD_CMD[@]}"

        if [ $? -eq 0 ]; then
            echo "✅ Build successful! Output: $OUTPUT_BINARY"
            if [ "$RUN_AFTER_BUILD" = true ]; then
                echo "🚀 Running built binary..."
                RUN_CMD=("$OUTPUT_BINARY" "${RUN_ARGS[@]}" -c "$CONFIG_FILE")
                # shellcheck disable=SC2145
                echo "Command: ${RUN_CMD[@]}"
                "${RUN_CMD[@]}"
            fi
        else
            echo "❌ Build failed!" >&2
            exit 1
        fi
        ;;
    *)
        echo "❌ Error: Unknown mode $MODE" >&2
        show_help
        ;;
esac