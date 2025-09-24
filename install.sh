#!/bin/bash

# ==============================================================================
# NetPilot Installation Script (v2.0 with Smart Diagnostics)
# ==============================================================================

# --- Configuration ---
BINARY_NAME="netpilot"
INSTALL_PATH="/usr/local/bin"
SERVICE_USER="netpilot"
SERVICE_FILE_PATH="/etc/systemd/system/netpilot.service"
# Key QoS kernel modules to check for
QOS_MODULES=("sch_cake" "sch_fq_codel")

# --- Helper Functions ---
print_info() { echo -e "\033[34m[INFO]\033[0m $1"; }
print_success() { echo -e "\033[32m[SUCCESS]\033[0m $1"; }
print_warning() { echo -e "\033[33m[WARNING]\033[0m $1"; }
print_error() { echo -e "\033[31m[ERROR]\033[0m $1"; exit 1; }
prompt_continue() {
    read -p "Do you wish to continue with the installation? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Installation aborted by user."
        exit 1
    fi
}

# --- Smart Diagnostics ---
check_qos_modules() {
    print_info "Checking for required QoS kernel modules..."
    local all_found=true
    
    for module in "${QOS_MODULES[@]}"; do
        if lsmod | grep -q "^$module"; then
            print_success "Module '$module' is already loaded."
        elif modinfo "$module" >/dev/null 2>&1; then
            print_warning "Module '$module' exists but is not loaded."
            read -p "Attempt to load it now? (Y/n) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]*$ ]]; then
                if sudo modprobe "$module"; then
                    print_success "Successfully loaded '$module'."
                else
                    print_error "Failed to load '$module'. This might indicate a deeper kernel issue."
                fi
            fi
        else
            print_warning "Module '$module' was not found on your system."
            all_found=false
        fi
    done

    if [ "$all_found" = false ]; then
        echo
        print_warning "One or more recommended QoS modules (like 'sch_cake') are missing."
        print_info "This means NetPilot's core functionality will not work correctly."
        print_info "This usually happens if your Linux kernel is too old or was compiled without these modules."
        print_info "For WSL2, running 'wsl --update' in PowerShell or compiling a custom kernel can solve this."
        echo
        prompt_continue
    else
        print_success "All recommended QoS modules are available!"
    fi
}


# --- Main Script Logic ---

# 1. Check for root privileges
if [ "$(id -u)" -ne 0 ]; then
    print_error "This script must be run as root. Please use sudo."
fi

# 2. Smart Diagnostics for QoS Modules
check_qos_modules

# 3. Check if the binary exists
# ... (The rest of the script is the same as before) ...
if [ ! -f "backend/$BINARY_NAME" ]; then
    print_warning "Backend binary not found at 'backend/$BINARY_NAME'."
    print_info "Attempting to build it now..."
    cd backend || exit 1
    go build -o $BINARY_NAME cmd/netpilot/main.go
    cd ..
    if [ ! -f "backend/$BINARY_NAME" ]; then
        print_error "Failed to build the binary. Please check for Go compilation errors."
    fi
fi

# 4. Create a dedicated user for the service
print_info "Creating a dedicated user '$SERVICE_USER' for the NetPilot service..."
if ! id -u "$SERVICE_USER" >/dev/null 2>&1; then
    useradd -r -s /bin/false "$SERVICE_USER"
    print_success "User '$SERVICE_USER' created."
else
    print_info "User '$SERVICE_USER' already exists. Skipping."
fi

# 5. Install the binary
print_info "Installing the '$BINARY_NAME' binary to '$INSTALL_PATH'..."
install -m 755 "backend/$BINARY_NAME" "$INSTALL_PATH"
print_success "Binary installed."

# 6. Create the systemd service file
print_info "Creating systemd service file at '$SERVICE_FILE_PATH'..."
cat << EOF > "$SERVICE_FILE_PATH"
[Unit]
Description=NetPilot API Server
After=network.target

[Service]
User=$SERVICE_USER
Group=$SERVICE_USER
AmbientCapabilities=CAP_NET_ADMIN
CapabilityBoundingSet=CAP_NET_ADMIN
NoNewPrivileges=true
Type=simple
ExecStart=$INSTALL_PATH/$BINARY_NAME
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF
print_success "Service file created."

# 7. Reload systemd, enable and start the service
print_info "Reloading systemd, enabling and starting the NetPilot service..."
systemctl daemon-reload
systemctl enable --now netpilot.service

# 8. Final check
print_info "Verifying service status..."
if systemctl is-active --quiet netpilot.service; then
    print_success "NetPilot service is now running!"
    echo "You can check its status anytime with: sudo systemctl status netpilot"
else
    print_error "The NetPilot service failed to start. Please check the logs with: sudo journalctl -u netpilot"
fi

echo
print_info "Installation complete."
