import axios from "axios";
import React, { useEffect, useRef } from "react";

interface VideoPlayerProps {
  videoEndpoint: string;
  fileName: string;
  onClose: () => void;
}

const VideoPlayer: React.FC<VideoPlayerProps> = ({
  videoEndpoint,
  fileName,
  onClose,
}) => {
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    const videoElement = videoRef.current;

    if (!videoElement) return;

    let currentChunkStart = 0;
    let nextChunkEnd = 0;
    let fetching = false;

    const fetchVideoChunk = async (rangeStart: number, rangeEnd: number) => {
      fetching = true;
      try {
        const headers: { Range: string } = {
          Range: `bytes=${rangeStart}-${rangeEnd}`,
        };

        const response = await fetch(`${videoEndpoint}/${fileName}`, {
          headers,
        });
        const videoBlob = await response.blob();
        const videoURL = URL.createObjectURL(videoBlob);

        videoElement.src = videoURL;
        videoElement.load();

        fetching = false;
      } catch (error) {
        console.error("Error fetching video chunk:", error);
        fetching = false;
      }
    };

    const handleTimeUpdate = () => {
      if (!videoElement) return;

      const bufferedLength = videoElement.buffered.length;
      if (bufferedLength === 0 || fetching) return;

      const bufferedEnd = videoElement.buffered.end(bufferedLength - 1);
      const currentTime = videoElement.currentTime;

      const remainingBuffer = bufferedEnd - currentTime;

      if (remainingBuffer < 10 && currentTime >= nextChunkEnd) {
        currentChunkStart = nextChunkEnd;
        nextChunkEnd = currentChunkStart + 500 * 1024; // 500KB
        fetchVideoChunk(currentChunkStart, nextChunkEnd);
      }
    };

    videoElement.addEventListener("timeupdate", handleTimeUpdate);

    return () => {
      videoElement.removeEventListener("timeupdate", handleTimeUpdate);
      handleLastVideoOpenData(videoElement.currentTime);
      setTimeout(() => {
        onClose();
      }, 1000);
    };
  }, [videoEndpoint, fileName]);

  const handleLastVideoOpenData = async (videoTime: number) => {
    const fileMinute = `${String(Math.floor(videoTime / 60)).padStart(
      2,
      "0"
    )}:${String(Math.floor(videoTime % 60)).padStart(2, "0")}`;

    console.log("updating video data...", fileName, fileMinute);
    axios
      .post("http://localhost:8080/last-access", {
        fileName,
        fileMinute,
      })
      .then(() => {
        console.debug("video data updated...");
      });
  };

  return (
    <div>
      <video
        ref={videoRef}
        width="889"
        height="560"
        controls
        preload="metadata"
      >
        <source src={`${videoEndpoint}/${fileName}`} type="video/mp4" />
        Your browser does not support the video tag.
      </video>
    </div>
  );
};

export default VideoPlayer;
