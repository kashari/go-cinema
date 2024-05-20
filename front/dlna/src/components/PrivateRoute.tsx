import { Navigate } from "react-router-dom";
import PropTypes from "prop-types";
import { ReactNode } from "react";

const PrivateRoute = ({ children }: { children: ReactNode }) => {
    const accessToken = localStorage.getItem("accessToken");

    return accessToken ? children : <Navigate to="/login" replace />;
};

PrivateRoute.propTypes = {
    children: PropTypes.node.isRequired,
};

export default PrivateRoute;