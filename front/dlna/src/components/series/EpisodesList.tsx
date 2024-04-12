import React, { useCallback, useEffect, useState } from "react";
import { Episode } from "../../types/series";
import axios, { AxiosProgressEvent } from "axios";
import { useParams } from "react-router-dom";
import play from "../../assets/play.svg";
import Modal from "../Modal";
import { SubmitHandler, useForm } from "react-hook-form";
import "../../App.css";
import FullScreenVideo from "../FullScreenVideo";

type EpisodeInputs = {
  File: File;
};

const EpisodesList: React.FC = () => {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<EpisodeInputs>();

  const [episodes, setEpisodes] = useState<Episode[]>([]);
  const [currentIndex, setCurrentIndex] = useState<number>(0);
  const [editModal, setEditModal] = useState<boolean>(false);
  const [episodeUploadProgress, setEpisodeUploadProgress] = useState<number>(0);
  const [videoModal, setVideoModal] = useState<boolean>(false);
  const [videoEndpoint, setVideoEndpoint] = useState<string>("");
  const [startOver, setStartOver] = useState<boolean>(false);
  const [currentEpisodePlaying, setCurrentEpisodePlaying] =
    useState<Episode | null>(null);

  const { id } = useParams();

  const onSerieSubmit: SubmitHandler<EpisodeInputs> = (data) => {
    const formData = new FormData();
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    formData.append("File", data.File[0]);

    axios
      .post(`http://192.168.3.150:8080/series/${id}/append`, formData, {
        method: "POST",
        headers: { "Content-Type": "multipart/form-data" },
        onUploadProgress: (progressEvent: AxiosProgressEvent) => {
          const percentCompleted = Math.round(
            (progressEvent.loaded * 100) / (progressEvent.total ?? 1)
          );
          setEpisodeUploadProgress(percentCompleted);
        },
      })
      .then((response) => {
        console.debug(response);
        setTimeout(() => {
          setEpisodeUploadProgress(0);
          handleCloseEditModal();
        }, 900);
      })
      .catch((error) => {
        console.error("Error:", error);
      });
  };

  const handleGetCurrentEpisodeIndex = useCallback(async () => {
    axios
      .get(`http://192.168.3.150:8080/series/${id}/current`)
      .then((response) => {
        console.debug(response);
        setCurrentIndex(response.data.index);
      });
  }, [id]);

  const handleSetCurrentEpisodeIndex = async (index: number) => {
    axios
      .post(`http://192.168.3.150:8080/series/${id}/current?index=${index}`)
      .then((response) => {
        console.debug(response);
      });
  };

  const handleOpenVideoModal = (index: number) => {
    setVideoEndpoint(
      `http://192.168.3.150:8080/video?file=${episodes[index - 1].Path}`
    );
    handleSetCurrentEpisodeIndex(index);
    setCurrentEpisodePlaying(episodes[index - 1]);
    setVideoModal(true);
  };

  const switchToNextEpisode = async () => {
    if (document.fullscreenEnabled) document.exitFullscreen();
    handleCloseVideoModal();

    const nextIndex = currentIndex + 1;

    if (nextIndex < episodes.length) {
      await handleSetCurrentEpisodeIndex(nextIndex);

      setTimeout(() => {
        handleOpenVideoModal(nextIndex);
      }, 3000);
    } else {
      console.log("No more episodes available");
    }
  };

  const handleCloseVideoModal = () => {
    handleGetCurrentEpisodeIndex();
    handleFetchEpisodes();
    setVideoModal(false);
  };

  const handleFetchEpisodes = useCallback(() => {
    axios
      .get(`http://192.168.3.150:8080/series/${id}/episodes`)
      .then((response) => {
        setEpisodes(response.data);
      });
  }, [id]);

  const handleCloseEditModal = () => {
    handleFetchEpisodes();
    setEditModal(false);
  };

  useEffect(() => {
    handleFetchEpisodes();
    handleGetCurrentEpisodeIndex();
  }, [handleFetchEpisodes, handleGetCurrentEpisodeIndex]);
  return (
    <>
      <div className="container-fluid  mt-5 mb-5">
        <div
          className="row mt-6 mb-6 p-6 border"
          onClick={() => {
            handleOpenVideoModal(currentIndex);
            setStartOver(false);
          }}
        >
          <img
            src={play}
            alt="play"
            height={150}
            width={150}
            title="Resume where you left at."
            style={{ cursor: "pointer" }}
          />
        </div>
        <div className="mt-6 mb-6">
          <h1 style={{ marginTop: "25px" }}>Episodes</h1>
          <span
            style={{ cursor: "pointer" }}
            onClick={() => {
              setEditModal(true);
            }}
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              height="18"
              width="16.5"
              viewBox="0 0 448 512"
            >
              <path
                fill="#FFD43B"
                d="M256 80c0-17.7-14.3-32-32-32s-32 14.3-32 32V224H48c-17.7 0-32 14.3-32 32s14.3 32 32 32H192V432c0 17.7 14.3 32 32 32s32-14.3 32-32V288H400c17.7 0 32-14.3 32-32s-14.3-32-32-32H256V80z"
              />
            </svg>
          </span>
        </div>
        <div className="row mt-5 mb-5">
          {episodes.map((episode) => (
            <div className="col-md-6 gy-4 col-sm-12" key={episode.ID}>
              <div
                className={`card mb-6 text-center ${
                  currentIndex === episode.EpisodeIndex ? "yellow-border" : ""
                }`}
              >
                <h3 className="text-center" style={{ marginTop: "25px" }}>
                  {episode.EpisodeIndex}
                </h3>
                <div
                  className="row"
                  style={{ marginTop: "15px" }}
                  onClick={() => {
                    handleOpenVideoModal(episode.EpisodeIndex);
                    setStartOver(true);
                  }}
                >
                  <img
                    src={play}
                    alt="play"
                    height={85}
                    width={85}
                    title="Resume where you left at."
                    style={{ cursor: "pointer" }}
                  />
                </div>
                <br />
                <small className="text-muted p-4 text-center">
                  {episode.ResumeAt}
                </small>

                <div className="card-body d-flex justify-content-around"></div>
              </div>
            </div>
          ))}
        </div>
        <Modal isOpen={editModal} onClose={handleCloseEditModal}>
          <div className="row">
            <form onSubmit={handleSubmit(onSerieSubmit)}>
              Add new episode
              <div className="row gy-3 mt-2">
                <div className="col-md-12 col-sm-12">
                  <input
                    required
                    type="file"
                    className="form-control mt-2"
                    {...register("File")}
                  />
                  {errors.File && (
                    <small className="text-danger">
                      Episode file is required
                    </small>
                  )}
                </div>
              </div>
              <div
                className="my-6"
                style={{ marginTop: "15px", marginBottom: "15px" }}
              >
                {episodeUploadProgress > 0 && (
                  <div className="flex items-center justify-center">
                    <div className="w-64">
                      <div className="progress">
                        <div
                          className="progress-bar progress-bar-striped bg-info"
                          role="progressbar"
                          style={{ width: `${episodeUploadProgress}%` }}
                          aria-valuenow={episodeUploadProgress}
                          aria-valuemin={0}
                          aria-valuemax={100}
                        >
                          {episodeUploadProgress}%
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>
              <div className="row mt-4">
                <div className="col-md-12 col-sm-12 mt-2">
                  <button type="submit" className="btn btn-primary">
                    Upload
                  </button>
                </div>
              </div>
            </form>
          </div>
        </Modal>
      </div>

      <FullScreenVideo
        isOpen={videoModal}
        episodeId={currentEpisodePlaying?.ID.toString() ?? ""}
        url={videoEndpoint}
        onClose={handleCloseVideoModal}
        leftAt={
          startOver ? "00:00" : currentEpisodePlaying?.ResumeAt ?? "00:00"
        }
        onEnded={switchToNextEpisode}
      />
    </>
  );
};

export default EpisodesList;
