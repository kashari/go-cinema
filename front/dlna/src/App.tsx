import React, { useCallback, useEffect, useState } from "react";
import "./App.css";
import { Button } from "react-bootstrap";
import axios from "axios";
import Modal from "./components/Modal";
import MoviePlayer from "./components/movies/MoviePlayer";
import { Link } from "react-router-dom";

type SerieData = {
  ID: string;
  Title: string;
  Path: string;
  Description: string;
  ResumeAt: string;
  CurrentIndex: string;
  Type: string;
  Episodes: Episode[];
};

type Episode = {
  EpisodeIndex: string;
  Path: string;
  ResumeAt: string;
};

type MovieData = {
  ID: string;
  Path: string;
  ResumeAt: string;
  Title: string;
  Description: string;
  Type: string;
};

const App: React.FC = () => {
  const [lastAccessMovie, setLastAccessMovie] = useState<MovieData | null>(
    null
  );
  const [lastAccessSerie, setLastAccessSerie] = useState<SerieData | null>(
    null
  );

  const [isMoviePlayerOpen, setIsMoviePlayerOpen] = useState<boolean>(false);

  const openMoviePlayer = () => {
    setIsMoviePlayerOpen(true);
  };

  const handleLastVideoOpenData = useCallback(async () => {
    axios.get("http://192.168.3.150:8080/left-at").then((response) => {
      const item = response.data;
      if (item.Type === "Movie") {
        setLastAccessMovie(item);
      } else {
        setLastAccessSerie(item);
      }
    });
  }, []);

  const closeMoviePlayer = () => {
    setIsMoviePlayerOpen(false);
    handleLastVideoOpenData();
  };

  useEffect(() => {
    handleLastVideoOpenData();
  }, [handleLastVideoOpenData]);
  return (
    <>
      <div className="container w-50" style={{ marginTop: "125px" }}>
        {lastAccessMovie && (
          <>
            <div>
              <h1 className="fw-boldest multicolor">
                Last time you left {lastAccessMovie.Title} at{" "}
                {lastAccessMovie.ResumeAt}
              </h1>
              <p className="text-muted">{lastAccessMovie.Description}</p>
            </div>
            <Button
              variant="outline-primary"
              size="lg"
              className="mt-5"
              onClick={openMoviePlayer}
            >
              Continue Watching
            </Button>
          </>
        )}

        {lastAccessSerie !== null && (
          <>
            <div>
              <h1 className="fw-boldest multicolor">
                Last time you left {lastAccessSerie.Title}{" "}
                {lastAccessSerie.CurrentIndex && (
                  <span> Episode {lastAccessSerie.CurrentIndex}</span>
                )}{" "}
                at{" "}
                {lastAccessSerie.Episodes &&
                  lastAccessSerie.Episodes[0].ResumeAt}
              </h1>
              <p className="text-muted">{lastAccessSerie?.Description}</p>
            </div>
            <Button variant="outline-primary" size="lg" className="mt-5">
              <Link
                to={`/series/${lastAccessSerie.ID}/episodes`}
                style={{ textDecoration: "none" }}
                state={{ title: lastAccessSerie.Title }}
              >
                Continue Watching
              </Link>
            </Button>
          </>
        )}

        <Modal isOpen={isMoviePlayerOpen} onClose={closeMoviePlayer}>
          {/* && (lastAccess as Movie).Path.includes("Movies") */}
          {lastAccessMovie ? (
            <MoviePlayer
              leftAt={lastAccessMovie.ResumeAt || "00:00"}
              movieId={lastAccessMovie?.ID || ""}
              videoEndpoint={"http://192.168.3.150:8080/video"}
              fileName={lastAccessMovie.Path || ""}
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
