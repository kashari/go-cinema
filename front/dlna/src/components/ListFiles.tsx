import React, { ChangeEvent, useCallback, useEffect, useState } from "react";
import { FileRow } from "../types/file";
import axios from "axios";
import folder from "../assets/folder.svg";
import Modal from "./Modal";
import Switch from "./Switch";
import VideoPlayer from "./VideoPlayer";
import { VideoStatus } from "../types/video-status";

const ListFiles: React.FC = () => {
  const [files, setFiles] = useState<FileRow[]>([]);
  const [file, setFile] = useState<File | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [fileUrl, setFileUrl] = useState<string>("");
  const [uploadProgress, setUploadProgress] = useState<number>(0);
  const [downloadProgress, setDownloadProgress] = useState<number>(0);
  const [hasDownloadInProgress, setHasDownloadInProgress] =
    useState<boolean>(false);
  const [currentFilePlaying, setCurrentFilePlaying] = useState<string>("");
  const [lastWatch, setLastWatch] = useState<VideoStatus | null>(null);

  const [videoOpen, setVideoOpen] = useState(false);

  useEffect(() => {
    fetchFiles();
    handleGetLastUpdate();
  }, []);

  const handleGetLastUpdate = async () => {
    const response = await axios.get("http://192.168.3.150:8080/last-access");
    setLastWatch(response.data);
  };

  const fetchFiles = async () => {
    // using axios to fetch files
    const response = await axios.get("http://192.168.3.150:8080/files");
    setFiles(response.data.files);
  };

  const handleDelete = (file: string) => {
    // Handle delete logic here
    if (!window.confirm(`Are you sure you want to delete ${file}?`)) {
      return;
    }
    console.debug(`Deleting file: ${file}`);
    axios.delete(`http://192.168.3.150:8080/files/${file}`);
    setFiles(files.filter((f) => f.Name !== file));
  };

  const handleUpload = async () => {
    if (!file) {
      console.error("No file selected");
      return;
    }

    const formData = new FormData();
    formData.append("file", file);

    try {
      await axios.post("http://192.168.3.150:8080/upload", formData, {
        headers: { "Content-Type": "multipart/form-data" },
        onUploadProgress: (progressEvent) => {
          const percentCompleted = Math.round(
            (progressEvent.loaded * 100) / (progressEvent.total ?? 1)
          );
          setUploadProgress(percentCompleted);
        },
      });
    } catch (error) {
      console.error("Error uploading file", error);
      return;
    }

    setTimeout(() => {
      setUploadProgress(0);
    }, 900);
    fetchFiles();
    // clear the file input
    setFile(null);
    const fileInput = document.querySelector(
      'input[type="file"]'
    ) as HTMLInputElement;
    if (fileInput) {
      fileInput.value = "";
    }
    console.log("File uploaded successfully");
  };

  const handleDownload = async () => {
    if (!fileUrl) {
      console.error("No file URL provided");
      return;
    }

    try {
      setHasDownloadInProgress(true);
      axios.get(`http://192.168.3.150:8080/download?url=${fileUrl}`);
    } catch (error) {
      console.error("Error downloading file", error);
      return;
    }
  };

  const handleVideoOpen = (fileName: string) => {
    setCurrentFilePlaying(fileName);
    setVideoOpen(true);
  };

  const handleVideoClose = () => {
    setCurrentFilePlaying("");
    setVideoOpen(false);
    handleGetLastUpdate();
  };

  const handleProgress = useCallback(async () => {
    axios
      .get(`http://192.168.3.150:8080/percentage?url=${fileUrl}`)
      .then((response) => {
        if (response.data.percentage === 100) {
          setTimeout(() => {
            setDownloadProgress(0);
            setHasDownloadInProgress(false);
            const textInput = document.querySelector(
              'input[type="text"]'
            ) as HTMLInputElement;
            textInput.value = "";
            const checkbox = document.querySelector(
              'input[type="checkbox"]'
            ) as HTMLInputElement;
            checkbox.checked = false;
            setIsModalOpen(false);
          }, 4000);
        }
        setDownloadProgress(response.data.percentage);
      });
  }, [fileUrl]);

  useEffect(() => {
    if (hasDownloadInProgress) {
      const interval = setInterval(() => {
        handleProgress();
      }, 1000);

      fetchFiles();
      return () => clearInterval(interval);
    }
  }, [hasDownloadInProgress, handleProgress]);

  const handleFileUri = (e: ChangeEvent<HTMLInputElement>) => {
    setFileUrl(e.target.value);
  };

  return (
    <>
      <div className="max-w-sm mx-auto bg-white rounded-xl shadow-md overflow-hidden sm:max-w-2xl">
        <div className="flex flex-col sm:flex-row">
          <div className="w-full sm:w-1/2 md:w-66 h-68 md:h-full">
            <img
              className="w-full max-w-md h-auto object-cover"
              src={folder}
              alt="Modern building architecture"
            />
          </div>
          <div className="p-8 sm:p-12">
            <div className="uppercase tracking-wide text-sm text-indigo-500 font-semibold mb-6">
              Upload a file
              <br />
              <br />
              <Switch
                onChange={(event) => {
                  if (event.target.checked) {
                    setIsModalOpen(true);
                  }
                }}
              />
            </div>
            <input
              type="file"
              onChange={(e) => setFile(e.target.files?.[0] || null)}
              className="block w-full text-sm text-slate-500
                  file:mr-4 file:py-2 file:px-4
                  file:rounded-full file:border-0
                  file:text-sm file:font-semibold
                  file:bg-violet-50 file:text-violet-700
                  hover:file:bg-violet-100 mb-8
                "
            />
            <button
              className="block w-full text-sm text-slate-500 bg-cyan-50 text-cyan-700 hover:bg-cyan-100 mb-8"
              onClick={handleUpload}
            >
              Upload
            </button>
            <p className="mt-2 text-slate-500">
              Use the form above to upload files to the server. When you select
              a file it will automatically upload to the server.
            </p>
          </div>
        </div>
      </div>
      <Modal
        isOpen={isModalOpen}
        onClose={() => {
          const checkbox = document.querySelector(
            'input[type="checkbox"]'
          ) as HTMLInputElement;
          checkbox.checked = false;
          setIsModalOpen(false);
        }}
      >
        <div className="p-8">
          <h2 className="text-xl font-bold mb-4">
            Download a file from the web.
          </h2>
          <input
            type="text"
            onChange={handleFileUri}
            className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
            placeholder="File URL"
            required
          />
          <br />
          {fileUrl}
          <br />
          <br />
          <button
            className="block w-full text-sm text-slate-500 bg-purple-50 text-purple-700 hover:bg-purple-100 mb-8"
            onClick={handleDownload}
          >
            Download
          </button>
        </div>
        <div className="my-6">
          {downloadProgress > 0 && (
            <div className="flex items-center justify-center">
              <div className="w-64 bg-gray-200">
                <div
                  className="bg-blue-500 text-xs leading-none py-1 text-center text-white"
                  style={{ width: `${downloadProgress}%` }}
                >
                  {downloadProgress}%
                </div>
              </div>
            </div>
          )}
        </div>
      </Modal>

      <div className="my-6">
        {uploadProgress > 0 && (
          <div className="flex items-center justify-center">
            <div className="w-64 bg-gray-200">
              <div
                className="bg-blue-500 text-xs leading-none py-1 text-center text-white"
                style={{ width: `${uploadProgress}%` }}
              >
                {uploadProgress}%
              </div>
            </div>
          </div>
        )}
      </div>

      <div className="py-16">
        <div className="mx-auto px-6 max-w-6xl text-gray-500">
          <div className="text-center">
            <h2 className="text-3xl text-gray-950 dark:text-white font-semibold">
              Last time you watched{" "}
              <code>{lastWatch?.file.split("/").pop()}</code> and left at{" "}
              <code>{lastWatch?.minute}</code>.
            </h2>
            <p className="mt-6 text-gray-700 dark:text-gray-300">
              Options, listed below, are available for each file. You can
              download, delete, or play video files.
            </p>
          </div>
          <div className="mt-12 grid sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {files.map((file) => (
              <div
                key={file.Name}
                className="relative group overflow-hidden p-8 rounded-xl bg-white border border-gray-200 dark:border-gray-800 dark:bg-gray-900"
              >
                <div
                  aria-hidden="true"
                  className="inset-0 absolute aspect-video border rounded-full -translate-y-1/2 group-hover:-translate-y-1/4 duration-300 bg-gradient-to-b from-blue-500 to-white dark:from-white dark:to-white blur-2xl opacity-25 dark:opacity-5 dark:group-hover:opacity-10"
                ></div>
                <div className="relative">
                  <div className="border border-yellow-500/10 flex relative *:relative *:size-6 *:m-auto size-12 rounded-lg dark:bg-gray-900 dark:border-white/15 before:rounded-[7px] before:absolute before:inset-0 before:border-t before:border-white before:from-yellow-100 dark:before:border-white/20 before:bg-gradient-to-b dark:before:from-white/10 dark:before:to-transparent before:shadow dark:before:shadow-gray-950">
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      width="0.73em"
                      height="1em"
                      viewBox="0 0 256 351"
                    >
                      <defs>
                        <filter
                          id="logosFirebase0"
                          width="200%"
                          height="200%"
                          x="-50%"
                          y="-50%"
                          filterUnits="objectBoundingBox"
                        >
                          <feGaussianBlur
                            in="SourceAlpha"
                            result="shadowBlurInner1"
                            stdDeviation="17.5"
                          ></feGaussianBlur>
                          <feOffset
                            in="shadowBlurInner1"
                            result="shadowOffsetInner1"
                          ></feOffset>
                          <feComposite
                            in="shadowOffsetInner1"
                            in2="SourceAlpha"
                            k2="-1"
                            k3="1"
                            operator="arithmetic"
                            result="shadowInnerInner1"
                          ></feComposite>
                          <feColorMatrix
                            in="shadowInnerInner1"
                            values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0.06 0"
                          ></feColorMatrix>
                        </filter>
                        <filter
                          id="logosFirebase1"
                          width="200%"
                          height="200%"
                          x="-50%"
                          y="-50%"
                          filterUnits="objectBoundingBox"
                        >
                          <feGaussianBlur
                            in="SourceAlpha"
                            result="shadowBlurInner1"
                            stdDeviation="3.5"
                          ></feGaussianBlur>
                          <feOffset
                            dx="1"
                            dy="-9"
                            in="shadowBlurInner1"
                            result="shadowOffsetInner1"
                          ></feOffset>
                          <feComposite
                            in="shadowOffsetInner1"
                            in2="SourceAlpha"
                            k2="-1"
                            k3="1"
                            operator="arithmetic"
                            result="shadowInnerInner1"
                          ></feComposite>
                          <feColorMatrix
                            in="shadowInnerInner1"
                            values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0.09 0"
                          ></feColorMatrix>
                        </filter>
                        <path
                          id="logosFirebase2"
                          d="m1.253 280.732l1.605-3.131l99.353-188.518l-44.15-83.475C54.392-1.283 45.074.474 43.87 8.188z"
                        ></path>
                        <path
                          id="logosFirebase3"
                          d="m134.417 148.974l32.039-32.812l-32.039-61.007c-3.042-5.791-10.433-6.398-13.443-.59l-17.705 34.109l-.53 1.744z"
                        ></path>
                      </defs>
                      <path
                        fill="#ffc24a"
                        d="m0 282.998l2.123-2.972L102.527 89.512l.212-2.017L58.48 4.358C54.77-2.606 44.33-.845 43.114 6.951z"
                      ></path>
                      <use
                        fill="#ffa712"
                        fillRule="evenodd"
                        href="#logosFirebase2"
                      ></use>
                      <use
                        filter="url(#logosFirebase0)"
                        href="#logosFirebase2"
                      ></use>
                      <path
                        fill="#f4bd62"
                        d="m135.005 150.38l32.955-33.75l-32.965-62.93c-3.129-5.957-11.866-5.975-14.962 0L102.42 87.287v2.86z"
                      ></path>
                      <use
                        fill="#ffa50e"
                        fillRule="evenodd"
                        href="#logosFirebase3"
                      ></use>
                      <use
                        filter="url(#logosFirebase1)"
                        href="#logosFirebase3"
                      ></use>
                      <path
                        fill="#f6820c"
                        d="m0 282.998l.962-.968l3.496-1.42l128.477-128l1.628-4.431l-32.05-61.074z"
                      ></path>
                      <path
                        fill="#fde068"
                        d="m139.121 347.551l116.275-64.847l-33.204-204.495c-1.039-6.398-8.888-8.927-13.468-4.34L0 282.998l115.608 64.548a24.126 24.126 0 0 0 23.513.005"
                      ></path>
                      <path
                        fill="#fcca3f"
                        d="M254.354 282.16L221.402 79.218c-1.03-6.35-7.558-8.977-12.103-4.424L1.29 282.6l114.339 63.908a23.943 23.943 0 0 0 23.334.006z"
                      ></path>
                      <path
                        fill="#eeab37"
                        d="M139.12 345.64a24.126 24.126 0 0 1-23.512-.005L.931 282.015l-.93.983l115.607 64.548a24.126 24.126 0 0 0 23.513.005l116.275-64.847l-.285-1.752z"
                      ></path>
                    </svg>
                  </div>
                  <div className="mt-6 pb-6 rounded-b-[--card-border-radius]">
                    <p className="text-gray-700 dark:text-gray-300">
                      {file.Name} ({file.Size})
                    </p>
                  </div>

                  <div className="flex gap-3 -mb-8 py-4 border-t border-gray-200 dark:border-gray-800">
                    <a
                      href={`http://192.168.3.150:8080/files/${file.Name}`}
                      target="_blank"
                      rel="noreferrer"
                      className="group rounded-xl disabled:border *:select-none [&>*:not(.sr-only)]:relative *:disabled:opacity-20 disabled:text-gray-950 disabled:border-gray-200 disabled:bg-gray-100 dark:disabled:border-gray-800/50 disabled:dark:bg-gray-900 dark:*:disabled:!text-white text-gray-950 bg-gray-100 hover:bg-gray-200/75 active:bg-gray-100 dark:text-white dark:bg-gray-500/10 dark:hover:bg-gray-500/15 dark:active:bg-gray-500/10 flex gap-1.5 items-center text-sm h-8 px-3.5 justify-center"
                    >
                      <span>Download</span>
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="1em"
                        height="1em"
                        viewBox="0 0 24 24"
                      >
                        <path
                          fill="none"
                          stroke="currentColor"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="m17 13l-5 5m0 0l-5-5m5 5V6"
                        ></path>
                      </svg>
                    </a>

                    <span
                      style={{ cursor: "pointer" }}
                      onClick={() => handleDelete(file.Name)}
                      className="group rounded-xl disabled:border *:select-none [&>*:not(.sr-only)]:relative *:disabled:opacity-20 disabled:text-gray-950 disabled:border-gray-200 disabled:bg-gray-100 dark:disabled:border-gray-800/50 disabled:dark:bg-gray-900 dark:*:disabled:!text-white text-gray-950 bg-gray-100 hover:bg-gray-200/75 active:bg-gray-100 dark:text-white dark:bg-gray-500/10 dark:hover:bg-gray-500/15 dark:active:bg-gray-500/10 flex gap-1.5 items-center text-sm h-8 px-3.5 justify-center"
                    >
                      <span>Delete</span>
                      &times;
                    </span>
                    {new RegExp(
                      [".mp4", ".MP4", ".mov", ".MOV", ".mpeg", ".MPEG"].join(
                        "|"
                      )
                    ).test(file.Name) && (
                      <span
                        onClick={() => handleVideoOpen(file.Name)}
                        style={{ cursor: "pointer" }}
                        className="group flex items-center rounded-xl disabled:border *:select-none [&>*:not(.sr-only)]:relative *:disabled:opacity-20 disabled:text-gray-950 disabled:border-gray-200 disabled:bg-gray-100 dark:disabled:border-gray-800/50 disabled:dark:bg-gray-900 dark:*:disabled:!text-white text-gray-950 bg-gray-100 hover:bg-gray-200/75 active:bg-gray-100 dark:text-white dark:bg-gray-500/10 dark:hover:bg-gray-500/15 dark:active:bg-gray-500/10 size-8 justify-center"
                      >
                        <span className="sr-only">Play</span>
                        <svg
                          xmlns="http://www.w3.org/2000/svg"
                          width={24}
                          height={24}
                          viewBox="0 0 384 512"
                        >
                          <path
                            fill="#FFD43B"
                            d="M73 39c-14.8-9.1-33.4-9.4-48.5-.9S0 62.6 0 80V432c0 17.4 9.4 33.4 24.5 41.9s33.7 8.1 48.5-.9L361 297c14.3-8.7 23-24.2 23-41s-8.7-32.2-23-41L73 39z"
                          />
                        </svg>
                      </span>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      <Modal isOpen={videoOpen} onClose={handleVideoClose}>
        <VideoPlayer
          videoEndpoint={"http://192.168.3.150:8080/video"}
          fileName={currentFilePlaying}
        />
      </Modal>
    </>
  );
};

export default ListFiles;
