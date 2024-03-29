# DLNA Frontend & File Handler

This project provides a user-friendly DLNA (Digital Living Network Alliance) frontend interface along with a backend file-handling service implemented in Go.

**Features**

- **DLNA browsing and interaction:** The frontend allows users to browse and interact with media accessible through a DLNA network.
- **File management:** Basic file management operations (e.g., view file details, limited modifications)
- **Customizable UI:** Built with React and Tailwind CSS for a modern look that can be easily adapted to your preferences.
- **Robust Backend:** The backend is powered by Golang for efficiency and reliability.

**Getting Started**

**Prerequisites**

- Node.js and npm (or yarn)
- Go (if you want to compile and run the backend)
- A running DLNA server on your network

**Frontend Installation**

1. Clone this repository: `git clone https://github.com/misenkashari/file-util.git`
2. Navigate to the frontend directory: `cd front/dlna`
3. Install dependencies: `npm install` or `yarn install`

**Backend Installation (Optional)**

1. Navigate to the root directory of the project
2. Build the Go backend: `go build`

**Running the Project**

1. **Build the frontend server:** From the `front/dlna` directory, run `npm run build` or `yarn build`.
2. **Start the frontend server:** Using `nginx` configure the `/etc/nginx/nginx.conf` as below:

   ```shell
   server {
   	    listen 80;
   	    server_name 192.168.*.*; # Or your server's IPv4 address

   	    root /home/kashari/dist/;
   	    index index.html;

   	    location / {
               try_files $uri $uri/ =404;
   	    }

   }
   ```

3. **Build the backend executable:** `.` directory, run `go build`.
4. **Start the Go backend:** From the root directory, run the compiled Go executable (e.g., `./dlna`).
5. **Register the Go backend as a service if you are in a Linux machine:** From the `/etc/systemd/system/` directory create a file named `dlna.service` and paste the below content:

   ```shell
    [Unit]
    Description=DLNA Backend Service
    After=network.target

    [Service]
    User=kashari
    WorkingDirectory=/home/kashari/dlna
    ExecStart=/home/kashari/dlna/dlna

    [Install]
    WantedBy=multi-user.target
   ```

   After this, use these commands:

   - `systemctl start dlna.service` to start the service.
   - `systemctl stop dlna.service` to stop the service.
   - `systemctl status dlna.service` to check health or status of the service.
   - `systemctl enable dlna.service` to enable the backend server at startup.

**Configuration**

- Modify frontend configuration (e.g., DLNA server address) in `front/dlna/src/config.ts`.
- Backend configuration can be managed in the `main.go` file or through environment variables.

**Project Structure**

- **front/dlna:** Contains the React frontend application
- **src:**
- **components:** Reusable UI components
- **types:** TypeScript definitions
- **App.tsx:** Main application entry point
- - **FileList.tsx:** The whole frontend basically.
- **index.html:** Base HTML template
- **io/file-handler.go:** Implements file handling logic in Go
- **main.go:** Entry point for the Go backend service
- **middlewares/logger.go:** Example middleware implementation for Go logging
- **LICENSE:** Contains the software license information

**Contributing**

We welcome contributions to this project! Please refer to the `CONTRIBUTING.md` file for guidelines on how to get involved.

**License**

This project is licensed under the [GNU GENERAL PUBLIC LICENSE] license. See the `LICENSE` file for details.
