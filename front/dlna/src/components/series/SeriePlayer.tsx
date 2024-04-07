import axios from "axios";
import React, { useCallback, useEffect, useRef, useState } from "react";

interface VideoPlayerProps {
  leftAt: string;
  movieId: string;
  videoEndpoint: string;
  fileName: string;
  onClose: () => void;
}

const SeriePlayer: React.FC<VideoPlayerProps> = ({
  leftAt,
  movieId,
  videoEndpoint,
  fileName,
  onClose,
}) => {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [playingIndex, setPlayingIndex] = useState<number>(0);

  const handleLastVideoOpenData = useCallback(
    (videoTime: number) => {
      const fileMinute = `${String(Math.floor(videoTime / 60)).padStart(
        2,
        "0"
      )}:${String(Math.floor(videoTime % 60)).padStart(2, "0")}`;

      console.debug("updating video data...", fileName, fileMinute);
      axios
        .post(`http://localhost:8080/last-access/${movieId}?time=${fileMinute}`)
        .then(() => {
          console.debug("video data updated...");
        });
    },
    [fileName, movieId]
  );

  const handleSkipToWhereYouLeft = useCallback(() => {
    if (videoRef.current && leftAt !== "00:00") {
      const [minutes, seconds] = leftAt.split(":");
      const timeInSeconds = parseInt(minutes) * 60 + parseInt(seconds);
      videoRef.current.currentTime = timeInSeconds;
    }
  }, [leftAt]);

  useEffect(() => {
    const videoElement = videoRef.current;

    if (!videoElement) return;

    setTimeout(() => {
      handleSkipToWhereYouLeft();
      videoElement.requestFullscreen();
      videoElement.play();
    }, 4000);

    const timeShifter = setInterval(() => {
      handleLastVideoOpenData(videoElement.currentTime);
    }, 60000);
    videoElement.addEventListener("ended", () => {
      setPlayingIndex((playingIndex) => playingIndex + 1);
      // get the next index file name
    });

    return () => {
      clearInterval(timeShifter);
      handleLastVideoOpenData(videoElement.currentTime);
      setTimeout(() => {
        onClose();
      }, 1000);
    };
  }, [
    videoEndpoint,
    fileName,
    playingIndex,
    handleLastVideoOpenData,
    onClose,
    handleSkipToWhereYouLeft,
  ]);

  return (
    <div>
      <video
        ref={videoRef}
        width="880"
        height="560"
        controls
        preload="metadata"
      >
        <source src={`${videoEndpoint}?file=${fileName}`} type="video/mp4" />
        Your browser does not support the video tag.
      </video>
    </div>
  );
};

export default SeriePlayer;
