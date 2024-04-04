import React, { useEffect, useRef } from "react";

interface VideoPlayerProps {
  videoEndpoint: string;
  fileName: string;
}

const VideoPlayer: React.FC<VideoPlayerProps> = ({
  videoEndpoint,
  fileName,
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

        const response = await fetch(
          `${videoEndpoint}/${fileName}?timePlaying=${String(
            Math.floor(videoElement.currentTime / 60)
          ).padStart(2, "0")}:${String(
            Math.floor(videoElement.currentTime % 60)
          ).padStart(2, "0")}`,
          {
            headers,
          }
        );
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
    };
  }, [videoEndpoint, fileName]);

  return (
    <div>
      <video
        ref={videoRef}
        width="640"
        height="360"
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
