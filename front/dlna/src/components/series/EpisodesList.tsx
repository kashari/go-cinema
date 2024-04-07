import React, { useCallback, useEffect, useState } from "react";
import { Episode } from "../../types/series";
import axios, { AxiosProgressEvent } from "axios";
import { useParams } from "react-router-dom";
import play from "../../assets/play.svg";
import Modal from "../Modal";
import { SubmitHandler, useForm } from "react-hook-form";

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
  const [editModal, setEditModal] = useState<boolean>(false);
  const [episodeUploadProgress, setEpisodeUploadProgress] = useState<number>(0);

  const { id } = useParams();

  const onSerieSubmit: SubmitHandler<EpisodeInputs> = (data) => {
    const formData = new FormData();
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    formData.append("File", data.File[0]);

    axios
      .post(`http://localhost:8080/series/${id}/append`, formData, {
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
        }, 900);
      })
      .catch((error) => {
        console.error("Error:", error);
      });
  };

  const handleFetchEpisodes = useCallback(() => {
    axios
      .get(`http://192.168.3.9:8080/series/${id}/episodes`)
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
  }, [handleFetchEpisodes]);
  return (
    <div className="container mt-6 mb-6">
      <div className="row mt-6 mb-6 p-6 border">
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
      <div className="row">
        {episodes.map((episode) => (
          <div className="col-md-3 gy-4 col-sm-12" key={episode.ID}>
            <div className="card mb-6">
              <h3 className="text-center" style={{ marginTop: "25px" }}>
                {episode.EpisodeIndex}
              </h3>
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
  );
};

export default EpisodesList;
