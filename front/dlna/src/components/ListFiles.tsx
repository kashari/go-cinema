import React, { ChangeEvent, useCallback, useEffect, useState } from "react";
import { FileRow } from "../types/file";
import axios from "axios";
import folder from "../assets/folder.svg";
import copy from "../assets/paperclip.svg";
import Modal from "./Modal";
import Switch from "./Switch";

const ListFiles: React.FC = () => {
  const [files, setFiles] = useState<FileRow[]>([]);
  const [file, setFile] = useState<File | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [fileUrl, setFileUrl] = useState<string>("");
  const [uploadProgress, setUploadProgress] = useState<number>(0);
  const [downloadProgress, setDownloadProgress] = useState<number>(0);
  const [hasDownloadInProgress, setHasDownloadInProgress] =
    useState<boolean>(false);

  useEffect(() => {
    fetchFiles();
  }, []);

  const fetchFiles = async () => {
    // using axios to fetch files
    const response = await axios.get("http://localhost:8080/files");
    setFiles(response.data.files);
  };

  const handleDelete = (file: string) => {
    // Handle delete logic here
    if (!window.confirm(`Are you sure you want to delete ${file}?`)) {
      return;
    }
    console.debug(`Deleting file: ${file}`);
    axios.delete(`http://localhost:8080/files/${file}`);
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
      await axios.post("http://localhost:8080/upload", formData, {
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
      axios.get(`http://localhost:8080/download?url=${fileUrl}`);
    } catch (error) {
      console.error("Error downloading file", error);
      return;
    }
  };

  const handleProgress = useCallback(async () => {
    axios
      .get(`http://localhost:8080/percentage?url=${fileUrl}`)
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

      <div className="md:flex mt-32 mx-auto justify-center">
        <ul role="list" className="divide-y divide-gray-100">
          {files.map((file) => (
            <li key={file.Name} className="py-4 flex">
              <img src={copy} alt="copy" className="h-10 w-10" />
              <div className="ml-3 flex flex-col">
                <p className="text-sm font-medium text-gray-900">{file.Name}</p>
                <p className="text-sm text-gray-500">{file.Size}</p>
                <div className="flex mt-2 gap-2">
                  <button
                    onClick={() => handleDelete(file.Name)}
                    className="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded-md shadow-sm text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                  >
                    Delete
                  </button>
                  <a
                    href={`http://localhost:8080/files/${file.Name}`}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                  >
                    <span className="ml-2">Download</span>
                  </a>
                </div>
              </div>
            </li>
          ))}
        </ul>
      </div>
    </>
  );
};

export default ListFiles;
