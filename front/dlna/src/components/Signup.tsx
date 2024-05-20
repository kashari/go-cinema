import React, { FormEvent, useState } from "react";
import axios from "../utils/axios";
import { Link } from "react-router-dom";

const Signup: React.FC = () => {

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [username, setUsername] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    try {
      await axios.post("/register", { username, email, password });
      window.location.replace("/login");
    } catch (err) {
      console.error(err);
      setError("Something went wrong. Please try again.")
    }
  };

    return (
        <div className="container">
            <div className="row mt-5">
                <div className="col-md-6 offset-md-3 mt-5">
                <h3 className="m-3">Signup</h3>
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
                    <label htmlFor="email">Email</label>
                    <input
                        type="email"
                        className="form-control"
                        id="email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
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
                    Signup
                    </button>
                    <Link to="/login" className="btn btn-link">Back</Link>
                </form>
                </div>
            </div>
        </div>
    );
};

export default Signup;