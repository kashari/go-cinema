import axios from "axios";

const axiosInstance = axios.create({
  baseURL: "http://192.168.3.150:8080",
});

axiosInstance.interceptors.request.use(
  (config) => {
    const accessToken = localStorage.getItem("accessToken");
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

axiosInstance.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    if (error.response.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      const refreshToken = localStorage.getItem("refreshToken");
      if (refreshToken) {
        try {
          const { data } = await axiosInstance.post("/refresh-token", {
            refreshToken,
          });
          localStorage.setItem("accessToken", data.access_token);
          originalRequest.headers.Authorization = `Bearer ${data.acccess_token}`;
          return axiosInstance(originalRequest);
        } catch (err) {
          localStorage.removeItem("accessToken");
          localStorage.removeItem("refreshToken");
          window.location.replace("/login");
        }
      } else {
        localStorage.removeItem("accessToken");
        window.location.replace("/login");
      }
    }
    return Promise.reject(error);
  }
);

export default axiosInstance;