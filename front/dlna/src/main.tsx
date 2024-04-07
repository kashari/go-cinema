import React from "react";
import ReactDOM from "react-dom/client";
import "bootstrap/dist/css/bootstrap.min.css";
import { RouterProvider, createBrowserRouter, Outlet } from "react-router-dom";
import NavigationPanel from "./components/NavigationPanel";
import App from "./App";
import Management from "./components/Management";
import SerieList from "./components/series/SerieList";
import MovieList from "./components/movies/MovieList";

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
    path: "/",
    element: <RootLayout />,
    children: [
      { path: "", element: <App /> },
      { path: "series", element: <SerieList /> },
      { path: "movies", element: <MovieList /> },
      { path: "management", element: <Management /> },
    ],
  },
]);

ReactDOM.createRoot(document.getElementById("root")!).render(
  //<React.StrictMode>
  <RouterProvider router={router} />
  //</React.StrictMode>
);
