import React, { FormEvent, useState } from "react";
import axios from "../utils/axios";
import { Link } from "react-router-dom";

const Login: React.FC = () => {
    const [username, setUsername] = useState<string>("");
    const [password, setPassword] = useState<string>("");
    const [error, setError] = useState<string>("");

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault();
        try {
          const { data } = await axios.post("/login", { username, password });
          localStorage.setItem("accessToken", data.access_token);
          localStorage.setItem("refreshToken", data.refresh_token);
          window.location.replace("/");
        } catch (err) {
          console.error(err);
            setError("Invalid username or password");
        }
      };
        
  return (
    <div className="container">
        <div className="row mt-5">
            <div className="col-md-6 offset-md-3 mt-5">
            <h3 className="m-3">Login</h3>
            <form onSubmit={(e: FormEvent<HTMLFormElement>) => handleSubmit(e)}>
                <div className="form-group m-3">
                <label htmlFor="username">Username</label>
                <input
                    type="text"
                    className="form-control"
                    id="username"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                />
                </div>
                <div className="form-group m-3">
                <label htmlFor="password">Password</label>
                <input
                    type="password"
                    className="form-control"
                    id="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                />
                </div>
                {error && <div className="alert alert-danger">{error}</div>}
                <button type="submit" className="btn btn-primary m-3">
                Login
                </button>
                No account? <Link to="/signup" className="btn btn-link">Signup</Link>
            </form>
            </div>
        </div>
    </div>
  );
};

export default Login;