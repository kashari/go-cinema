import React from "react";
import ReactDOM from "react-dom/client";
import "bootstrap/dist/css/bootstrap.min.css";
import { RouterProvider, createBrowserRouter, Outlet } from "react-router-dom";
import NavigationPanel from "./components/NavigationPanel";
import App from "./App";
import Management from "./components/Management";
import SerieList from "./components/series/SerieList";
import MovieList from "./components/movies/MovieList";
import EpisodesList from "./components/series/EpisodesList";
import PrivateRoute from "./components/PrivateRoute";
import Login from "./components/Login";
import Signup from "./components/Signup";

export const RootLayout: React.FC = () => {
  return (
    <div className="container">
      <NavigationPanel />
      <Outlet />
    </div>
  );
};

const router = createBrowserRouter([
  {
    path: "/login",
    element: <Login />,
  },
  {
    path: "/signup",
    element: <Signup />,
  },
  {
    path: "/",
    element: <PrivateRoute><RootLayout /></PrivateRoute>,
    children: [
      { path: "", element: <PrivateRoute><App /></PrivateRoute>},
      { path: "series", element: <PrivateRoute><SerieList /></PrivateRoute>},
      { path: "movies", element: <PrivateRoute><MovieList /></PrivateRoute>},
      { path: "management", element: <PrivateRoute><Management /></PrivateRoute>},
      { path: "series/:id/episodes", element: <PrivateRoute><EpisodesList /></PrivateRoute>},
    ],
  },
]);

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
  <RouterProvider router={router} />
  </React.StrictMode>
);
