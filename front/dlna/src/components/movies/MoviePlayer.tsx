import axios from "../../utils/axios";
import React, { useCallback, useEffect, useRef } from "react";

interface VideoPlayerProps {
  leftAt: string;
  movieId: string;
  videoEndpoint: string;
  fileName: string;
  onClose: () => void;
}

const MoviePlayer: React.FC<VideoPlayerProps> = ({
  leftAt,
  movieId,
  videoEndpoint,
  fileName,
  onClose,
}) => {
  const videoRef = useRef<HTMLVideoElement>(null);

  const handleLastVideoOpenData = useCallback(
    (videoTime: number) => {
      const fileMinute = `${String(Math.floor(videoTime / 60)).padStart(
        2,
        "0"
      )}:${String(Math.floor(videoTime % 60)).padStart(2, "0")}`;

      console.debug("updating video data...", fileName, fileMinute);
      axios
        .post(
          `http://192.168.3.200:9090/last-access/${movieId}?time=${fileMinute}`
        )
        .then(() => {
          console.debug("video data updated...");
        });
    },
    [fileName, movieId]
  );

  const handleBackGoroutine = useCallback(async () => {
    const response = await axios.post(
      `http://192.168.3.200:9090/start-cronos?interval=@every 10m`
    );

    if (response.status === 200) {
      console.debug("Goroutine started successfully");
    } else {
      console.error("Failed to start goroutine");
    }
  }, []);

  const handleBackGoroutineStop = useCallback(async () => {
      const response = await axios.post(
        `http://192.168.3.200:9090/stop-cronos`
      );
  
      if (response.status === 200) {
        console.debug("Goroutine stopped successfully");
      } else {
        console.error("Failed to stop goroutine");
      }
    }, []);

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

    handleBackGoroutine();

    setTimeout(() => {
      handleSkipToWhereYouLeft();
      videoElement.requestFullscreen();
      videoElement.play();
    }, 4000);

    const timeShifter = setInterval(() => {
      handleLastVideoOpenData(videoElement.currentTime);
      handleBackGoroutine();
    }, 60000);

    return () => {
      clearInterval(timeShifter);
      handleBackGoroutineStop();
      handleLastVideoOpenData(videoElement.currentTime);
      setTimeout(() => {
        onClose();
      }, 1000);
    };
  }, [
    videoEndpoint,
    fileName,
    handleLastVideoOpenData,
    onClose,
    handleSkipToWhereYouLeft,
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
        <source src={`${videoEndpoint}?file=${fileName}`} type="video/mp4" />
        Your browser does not support the video tag.
      </video>
    </div>
  );
};

export default MoviePlayer;
