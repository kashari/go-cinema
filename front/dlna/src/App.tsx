import React, { useEffect, useState } from "react";
import "./App.css";
import { Button } from "react-bootstrap";
import { Movie } from "./types/movie";
import axios from "axios";
import { Series } from "./types/series";
import Modal from "./components/Modal";
import MoviePlayer from "./components/movies/MoviePlayer";

const App: React.FC = () => {
  const [lastAccess, setLastAccess] = useState<Movie | Series | null>(null);
  const [isMoviePlayerOpen, setIsMoviePlayerOpen] = useState<boolean>(false);

  const openMoviePlayer = () => {
    setIsMoviePlayerOpen(true);
  };

  const handleLastVideoOpenData = async () => {
    axios.get("http://localhost:8080/left-at").then((response) => {
      setLastAccess(response.data);
    });
  };

  const closeMoviePlayer = () => {
    setIsMoviePlayerOpen(false);
    handleLastVideoOpenData();
  };

  useEffect(() => {
    handleLastVideoOpenData();
  }, []);
  return (
    <>
      <div className="container">
        {lastAccess ? (
          <>
            <h1 className="fw-boldest multicolor">
              Last time you left {lastAccess?.Title}{" "}
              {(lastAccess as Series).CurrentIndex && (
                <span>, Episode {(lastAccess as Series).CurrentIndex}</span>
              )}{" "}
              at {(lastAccess as Movie)?.ResumeAt}
            </h1>
            <p className="text-muted">{lastAccess?.Description}</p>
            <Button
              variant="outline-primary"
              size="lg"
              className="mt-5"
              onClick={openMoviePlayer}
            >
              Continue Watching
            </Button>
          </>
        ) : (
          <h1 className="fw-boldest multicolor">
            No data found for last time watch by you, choose a movie or a serie
            and start watching now.
          </h1>
        )}
        <Modal isOpen={isMoviePlayerOpen} onClose={closeMoviePlayer}>
          {/* && (lastAccess as Movie).Path.includes("Movies") */}
          {lastAccess ? (
            <MoviePlayer
              leftAt={(lastAccess as Movie)?.ResumeAt || "00:00"}
              movieId={lastAccess?.ID || ""}
              videoEndpoint={"http://localhost:8080/video"}
              fileName={(lastAccess as Movie).Path || ""}
              onClose={closeMoviePlayer}
            />
          ) : (
            <>Series coming soon</>
          )}
        </Modal>
      </div>
    </>
  );
};

export default App;
