import React, { useEffect, useState } from "react";
import { Movie } from "../../types/movie";
import axios from "../../utils/axios";
import Modal from "../Modal";
import { SubmitHandler, useForm } from "react-hook-form";
import { Series } from "../../types/series";
import { Link } from "react-router-dom";

type SerieInputs = {
  Title: string;
  Description: string;
};

const SerieList: React.FC = () => {
  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors },
  } = useForm<SerieInputs>();

  const onSerieSubmit: SubmitHandler<SerieInputs> = (data) => {
    axios
      .put(`http://192.168.3.200:9090/series/${updatingSerie?.ID}`, data)
      .then((response) => {
        console.debug(response);
      })
      .catch((error) => {
        console.error("Error:", error);
      });
  };

  const [series, setSeries] = useState<Series[]>([]);
  const [updatingSerie, setUpdatingSerie] = useState<Movie | null>(null);
  const [editModal, setEditModal] = useState<boolean>(false);

  const handleOpenEditModal = async (id: string) => {
    const response = await axios.get(`http://192.168.3.200:9090/series/${id}`);
    setUpdatingSerie(response.data);
    setEditModal(true);

    setTimeout(() => {
      if (updatingSerie) {
        setValue("Description", updatingSerie.Description);
        setValue("Title", updatingSerie.Title);
      }
    }, 300);
  };

  const handleCloseEditModal = () => {
    reset();
    setEditModal(false);
    handleFetchSeries();
  };

  const handleFetchSeries = () => {
    axios.get("http://192.168.3.200:9090/series").then((response) => {
      setSeries(response.data);
    });
  };

  const handleDelete = (id: string) => {
    axios
      .delete(`http://192.168.3.200:9090/series/${id}`)
      .then((response) => {
        console.debug(response);
        handleFetchSeries();
      })
      .catch((error) => {
        console.error(error);
      });
  };

  useEffect(() => {
    handleFetchSeries();
    if (updatingSerie) {
      setValue("Description", updatingSerie.Description);
      setValue("Title", updatingSerie.Title);
    }
  }, [setValue, updatingSerie]);
  return (
    <div className="container mt-4 mb-4">
      <div className="row">
        {series.map((serie) => (
          <div className="col-md-6 gy-4 col-sm-12" key={serie.ID}>
            <div className="card mb-6">
              <Link
                to={`/series/${serie.ID}/episodes`}
                state={{ currentIndex: serie.CurrentIndex, title: serie.Title }}
                style={{
                  textDecoration: "none",
                  color: "#89CFF0",
                }}
              >
                <h6 className="text-center mt-5" style={{ fontSize: "45px" }}>
                  {serie.Title}
                </h6>
              </Link>
              <br />
              <small className="text-muted p-4 text-center">
                {serie.Description}
              </small>

              <div className="card-body d-flex justify-content-around">
                <span
                  style={{ cursor: "pointer" }}
                  onClick={() => {
                    if (confirm(`Delete serie ${serie.Title} ?`))
                      handleDelete(serie.ID);
                  }}
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    height="22"
                    width="20"
                    viewBox="0 0 448 512"
                  >
                    <path
                      fill="#FFD43B"
                      d="M135.2 17.7C140.6 6.8 151.7 0 163.8 0H284.2c12.1 0 23.2 6.8 28.6 17.7L320 32h96c17.7 0 32 14.3 32 32s-14.3 32-32 32H32C14.3 96 0 81.7 0 64S14.3 32 32 32h96l7.2-14.3zM32 128H416V448c0 35.3-28.7 64-64 64H96c-35.3 0-64-28.7-64-64V128zm96 64c-8.8 0-16 7.2-16 16V432c0 8.8 7.2 16 16 16s16-7.2 16-16V208c0-8.8-7.2-16-16-16zm96 0c-8.8 0-16 7.2-16 16V432c0 8.8 7.2 16 16 16s16-7.2 16-16V208c0-8.8-7.2-16-16-16zm96 0c-8.8 0-16 7.2-16 16V432c0 8.8 7.2 16 16 16s16-7.2 16-16V208c0-8.8-7.2-16-16-16z"
                    />
                  </svg>
                </span>

                <span
                  style={{ cursor: "pointer" }}
                  onClick={() => {
                    handleOpenEditModal(serie.ID);
                  }}
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    height="22"
                    width="22"
                    viewBox="0 0 512 512"
                  >
                    <path
                      fill="#FFD43B"
                      d="M441 58.9L453.1 71c9.4 9.4 9.4 24.6 0 33.9L424 134.1 377.9 88 407 58.9c9.4-9.4 24.6-9.4 33.9 0zM209.8 256.2L344 121.9 390.1 168 255.8 302.2c-2.9 2.9-6.5 5-10.4 6.1l-58.5 16.7 16.7-58.5c1.1-3.9 3.2-7.5 6.1-10.4zM373.1 25L175.8 222.2c-8.7 8.7-15 19.4-18.3 31.1l-28.6 100c-2.4 8.4-.1 17.4 6.1 23.6s15.2 8.5 23.6 6.1l100-28.6c11.8-3.4 22.5-9.7 31.1-18.3L487 138.9c28.1-28.1 28.1-73.7 0-101.8L474.9 25C446.8-3.1 401.2-3.1 373.1 25zM88 64C39.4 64 0 103.4 0 152V424c0 48.6 39.4 88 88 88H360c48.6 0 88-39.4 88-88V312c0-13.3-10.7-24-24-24s-24 10.7-24 24V424c0 22.1-17.9 40-40 40H88c-22.1 0-40-17.9-40-40V152c0-22.1 17.9-40 40-40H200c13.3 0 24-10.7 24-24s-10.7-24-24-24H88z"
                    />
                  </svg>
                </span>
              </div>
            </div>
          </div>
        ))}
      </div>

      <Modal isOpen={editModal} onClose={handleCloseEditModal}>
        <div className="row">
          <form onSubmit={handleSubmit(onSerieSubmit)}>
            Update Serie <code>{updatingSerie?.Title}</code>
            <div className="row gy-3 mt-2">
              <div className="col-md-6 col-sm-12 mt-2">
                <input
                  type="text"
                  className="form-control"
                  placeholder="Movie title"
                  {...register("Title")}
                />
                {errors.Title && (
                  <small className="text-danger">Serie title is required</small>
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
                    Serie description is required
                  </small>
                )}
              </div>
            </div>
            <div className="row mt-4">
              <div className="col-md-12 col-sm-12 mt-2">
                <button type="submit" className="btn btn-primary">
                  Update
                </button>
              </div>
            </div>
          </form>
        </div>
      </Modal>
    </div>
  );
};

export default SerieList;
