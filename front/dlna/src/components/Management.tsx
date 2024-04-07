import React, { useState } from "react";
import "../App.css";

import { useForm, SubmitHandler } from "react-hook-form";
import axios, { AxiosProgressEvent } from "axios";

type MovieInputs = {
  Title: string;
  Description: string;
  File: File;
};

type SerieInputs = {
  Title: string;
  Description: string;
};

const Management: React.FC = () => {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<MovieInputs>();

  const {
    register: serieRegister,
    handleSubmit: serieSubmit,
    reset: serieReset,
    formState: { errors: serieErrors },
  } = useForm<SerieInputs>();

  const [movieUploadProgress, setMovieUploadProgress] = useState<number>(0);

  const onMovieSubmit: SubmitHandler<MovieInputs> = (data) => {
    const formData = new FormData();
    console.log(data.File);
    formData.append("Title", data.Title);
    formData.append("Description", data.Description);
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    formData.append("File", data.File[0]);
    axios
      .post("http://localhost:8080/movies", formData, {
        method: "POST",
        headers: { "Content-Type": "multipart/form-data" },
        onUploadProgress: (progressEvent: AxiosProgressEvent) => {
          const percentCompleted = Math.round(
            (progressEvent.loaded * 100) / (progressEvent.total ?? 1)
          );
          setMovieUploadProgress(percentCompleted);
        },
      })
      .then((response) => {
        reset();
        console.debug(response.data);
        setTimeout(() => {
          setMovieUploadProgress(0);
        }, 900);
      })
      .catch((error) => {
        console.error("Error:", error);
      });
  };

  const onSerieSubmit: SubmitHandler<SerieInputs> = (data) => {
    axios
      .post("http://localhost:8080/series", {
        body: data,
      })
      .then((response) => {
        serieReset();
        console.debug(response.data);
      });
  };

  return (
    <div className="container mt-4 mb-4">
      <h1 className="multicolor">Management options</h1>
      <h6>Here you can add movies and/or series</h6>
      <div className="row p-4 border shadow-sm">
        <div className="row gx-1">
          <div className="col-md-6 col-sm-12 p-4">
            Add a new Movie
            <form onSubmit={handleSubmit(onMovieSubmit)}>
              <div className="row gy-3 mt-2">
                <div className="col-md-6 col-sm-12 mt-2">
                  <input
                    type="text"
                    className="form-control"
                    placeholder="Movie title"
                    {...register("Title")}
                  />
                  {errors.Title && (
                    <small className="text-danger">
                      Movie title is required
                    </small>
                  )}
                </div>
                <div className="col-md-6 col-sm-12 mt-2">
                  <input
                    type="text"
                    className="form-control"
                    placeholder="Movie description"
                    {...register("Description")}
                  />
                  {errors.Description && (
                    <small className="text-danger">
                      Movie description is required
                    </small>
                  )}
                </div>
              </div>
              <div className="row mb-4">
                <div className="col-md-12 col-sm-12">
                  <input
                    type="file"
                    className="form-control mt-2"
                    {...register("File")}
                  />
                  {errors.File && (
                    <small className="text-danger">
                      Movie file is required
                    </small>
                  )}
                </div>
              </div>
              <div className="my-6">
                {movieUploadProgress > 0 && (
                  <div className="flex items-center justify-center">
                    <div className="w-64">
                      <div className="progress">
                        <div
                          className="progress-bar progress-bar-striped bg-info"
                          role="progressbar"
                          style={{ width: `${movieUploadProgress}%` }}
                          aria-valuenow={movieUploadProgress}
                          aria-valuemin={0}
                          aria-valuemax={100}
                        >
                          {movieUploadProgress}%
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>
              <div className="row mt-4">
                <div className="col-md-12 col-sm-12 mt-2">
                  <button type="submit" className="btn btn-primary">
                    Add Movie
                  </button>
                </div>
              </div>
            </form>
          </div>
          <div className="col-md-6 col-sm-12 p-4">
            Add a new serie
            <form onSubmit={serieSubmit(onSerieSubmit)}>
              <div className="row">
                <div className="col-md-12 col-sm-12 mt-2">
                  <input
                    type="text"
                    className="form-control"
                    placeholder="Serie title"
                    {...serieRegister("Title")}
                  />
                  {serieErrors.Title && (
                    <small className="text-danger">
                      Series title is required
                    </small>
                  )}
                </div>
              </div>
              <div className="row mb-6">
                <div className="col-md-12 col-sm-12 mt-2">
                  <input
                    type="text"
                    className="form-control"
                    placeholder="Serie description"
                    {...serieRegister("Description")}
                  />
                  {serieErrors.Description && (
                    <small className="text-danger">
                      Series description is required
                    </small>
                  )}
                </div>
              </div>
              <div className="row mt-4">
                <div className="col-md-12 col-sm-12 mt-2">
                  <button type="submit" className="btn btn-primary">
                    Add Serie
                  </button>
                </div>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Management;
