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

    const timeShifter = setInterval(() => {
      handleLastVideoOpenData(videoElement.currentTime);
    }, 60000);

    let currentChunkStart = 0;
    let nextChunkEnd = 0;
    let fetching = false;

    const fetchVideoChunk = async (rangeStart: number, rangeEnd: number) => {
      fetching = true;
      try {
        const headers: { Range: string } = {
          Range: `bytes=${rangeStart}-${rangeEnd}`,
        };

        const mediaSource = new MediaSource();
        const videoBuffer = mediaSource.addSourceBuffer(
          "video/mp4; codecs=avc1.42E01E,mp4a.40.2"
        );

        videoBuffer.mode = "sequence";

        await axios
          .get(`${videoEndpoint}/${fileName}`, {
            headers,
            responseType: "blob",
          })
          .then((response: any) => {
            const videoBlob = response.arrayBuffer();
            videoBuffer.appendBuffer(videoBlob);
          });

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
        nextChunkEnd = currentChunkStart + 300 * 1024 + 1; // 300 KB
        fetchVideoChunk(currentChunkStart, nextChunkEnd);
      }
    };

    videoElement.addEventListener("timeupdate", handleTimeUpdate);

    return () => {
      videoElement.removeEventListener("timeupdate", handleTimeUpdate);
      clearInterval(timeShifter);
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
      .post("http://192.168.3.150:8080/last-access", {
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
        width="880"
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
