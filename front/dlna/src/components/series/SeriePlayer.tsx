import axios from "../../utils/axios";
import React, { useCallback, useEffect, useRef, useState } from "react";

interface VideoPlayerProps {
  leftAt: string;
  episodeId: string;
  videoEndpoint: string;
  onClose: () => void;
  onEnded?: () => void;
}

const SeriePlayer: React.FC<VideoPlayerProps> = ({
  leftAt,
  episodeId,
  videoEndpoint,
  onClose,
  onEnded,
}) => {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [fullScreenButton, setFullScreenButton] = useState<boolean>(false);

  const handleLastVideoOpenData = useCallback(
    (videoTime: number) => {
      const fileMinute = `${String(Math.floor(videoTime / 60)).padStart(
        2,
        "0"
      )}:${String(Math.floor(videoTime % 60)).padStart(2, "0")}`;

      console.debug("updating video data...", fileMinute);
      axios
        .post(
          `http://192.168.3.200:9090/episodes/${episodeId}/last-access?time=${fileMinute}`
        )
        .then(() => {
          console.debug("video data updated...");
        });
    },
    [episodeId]
  );

  const handleSkipToWhereYouLeft = useCallback(() => {
    if (videoRef.current && leftAt !== "00:00") {
      const [minutes, seconds] = leftAt.split(":");
      const timeInSeconds = parseInt(minutes) * 60 + parseInt(seconds);
      videoRef.current.currentTime = timeInSeconds;
    }
  }, [leftAt]);

  const handleBackGoroutine = async () => {
    const response = await axios.post(
      `http://192.168.3.200:9090/start-cronos?interval=@every-10m`
    );

    if (response.status === 200) {
      console.debug("Goroutine started successfully");
    } else {
      console.error("Failed to start goroutine");
    }
  }

  const handleBackGoroutineStop = async () => {
      const response = await axios.post(
        `http://192.168.3.200:9090/stop-cronos`
      );
  
      if (response.status === 200) {
        console.debug("Goroutine stopped successfully");
      } else {
        console.error("Failed to stop goroutine");
      }
    }

  useEffect(() => {
    const videoElement = videoRef.current;

    if (!videoElement) return;
    setFullScreenButton(true);

    // start the goroutine that restarts the backend service periodically
    handleBackGoroutine();

    setTimeout(() => {
      handleSkipToWhereYouLeft();
      const fsButton = document.getElementById("fs-button");
      fsButton?.setAttribute("aria-pressed", "true");
      setFullScreenButton(false);
      videoElement.play();
    }, 2000);

    const timeShifter = setInterval(() => {
      handleLastVideoOpenData(videoElement.currentTime);
    }, 60000);
    videoElement.addEventListener("ended", () => {
      handleLastVideoOpenData(videoElement.currentTime);
      onEnded && onEnded();
    });

    return () => {
      clearInterval(timeShifter);
      handleBackGoroutineStop();
      handleLastVideoOpenData(videoElement.currentTime);
      setTimeout(() => {
        onClose();
      }, 1000);
      videoElement.removeEventListener("ended", () => {
        handleLastVideoOpenData(videoElement.currentTime);
        onEnded && onEnded();
      });
    };
  }, [
    videoEndpoint,
    handleLastVideoOpenData,
    onClose,
    handleSkipToWhereYouLeft,
    onEnded,
  ]);

  return (
    <div>
      <video
        ref={videoRef}
        width="95%"
        height="95%"
        controls
        preload="metadata"
      >
        <source src={`${videoEndpoint}`} type="video/mp4" />
        Your browser does not support the video tag.
      </video>
      {fullScreenButton && (
        <button
          style={{ display: "none" }}
          id="fs-button"
          onClick={(e) => {
            e.preventDefault();
            const videoElement = document.querySelector("video");
            videoElement?.requestFullscreen();
          }}
        >
          fs
        </button>
      )}
    </div>
  );
};

export default SeriePlayer;
