import React, { useState, useRef, useCallback, useEffect } from "react";
import styled from "styled-components";
import "@fortawesome/fontawesome-free/css/all.css";
import axios from "axios";

const LeftIcon = styled.i<{ rotated: boolean }>`
  ${({ rotated }) => rotated && "transform: rotate(-85deg);"}
  transition: transform 0.3s ease;
`;

const RightIcon = styled.i<{ rotated: boolean }>`
  ${({ rotated }) => rotated && "transform: rotate(85deg);"}
  transition: transform 0.3s ease;
`;

const VideoContainer = styled.div`
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: black;
  display: flex;
  justify-content: center;
  align-items: center;
  flex-direction: column;
`;

const VideoPlayer = styled.video`
  width: 100%;
  height: 100%;
`;

const ControlsContainer = styled.div<{ showControls: boolean }>`
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  width: 70%;
  justify-content: space-between;
  opacity: ${({ showControls }) => (showControls ? 1 : 0)};
  transition: opacity 0.3s ease;
`;

const ProgressBarContainer = styled.div<{ showControls: boolean }>`
  width: 80%;
  height: 20px;
  position: absolute;
  bottom: 5px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 10px;
  opacity: ${({ showControls }) => (showControls ? 1 : 0)};
  transition: opacity 0.3s ease;
`;

const ProgressBar = styled.div<{ progress: number }>`
  width: 100%;
  height: 4px;
  background: linear-gradient(
    to right,
    red ${({ progress }) => progress}%,
    white ${({ progress }) => progress}%,
    white
  );
  border-radius: 2px;
  position: relative;
  cursor: pointer;
`;

const ProgressIndicator = styled.div<{ progress: number }>`
  width: 22px; /* slightly larger */
  height: 22px; /* slightly larger */
  background-color: white; /* white */
  cursor: pointer;
  border-radius: 50%;
  position: absolute;
  top: 50%;
  left: ${({ progress }) => `${progress}%`};
  transform: translate(-50%, -50%);
`;

const TimeInfo = styled.div<{ showControls: boolean }>`
  position: absolute;
  width: 77%;
  color: white;
  font-size: 14px;
  display: flex;
  justify-content: space-between;
  bottom: 3%;
  transition: opacity 0.3s ease;
  opacity: ${({ showControls }) => (showControls ? 1 : 0)};
`;

const ControlButton = styled.button`
  color: white;
  font-size: 100px;
  border: none;
  background-color: transparent;
  margin: 0 20px;
  cursor: pointer;
  transition: color 0.3s;

  &:hover {
    color: rgba(255, 255, 255, 0.7);
  }
`;

const PauseButton = styled.button`
  color: white;
  font-size: 170px;
  border: none;
  background-color: transparent;
  margin: 0 20px;
  cursor: pointer;
  transition: color 0.3s;

  &:hover {
    color: rgba(255, 255, 255, 0.7);
  }
`;

type FullScreenVideoProps = {
  url: string;
  isOpen: boolean;
  episodeId: string;
  leftAt: string;
  onClose: () => void;
  onEnded?: () => void;
};

const FullScreenVideo: React.FC<FullScreenVideoProps> = ({
  url,
  isOpen,
  episodeId,
  leftAt,
  onClose,
  onEnded,
}) => {
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState<number>(0);
  const [duration, setDuration] = useState<number>(0);
  const [showControls, setShowControls] = useState<boolean>(true);

  const videoRef = useRef<HTMLVideoElement>(null);
  const controlsTimeoutRef = useRef<ReturnType<typeof setTimeout>>();

  const [isRotateLeftClicked, setIsRotateLeftClicked] =
    useState<boolean>(false);
  const [isRotateRightClicked, setIsRotateRightClicked] =
    useState<boolean>(false);

  const handleRotateLeftClick = () => {
    setIsRotateLeftClicked(!isRotateLeftClicked);
    setIsRotateRightClicked(false);
    setTimeout(() => {
      setIsRotateLeftClicked(false);
    }, 300);
    handleSeek(-15);
  };

  const handleRotateRightClick = () => {
    setIsRotateRightClicked(!isRotateRightClicked);
    setIsRotateLeftClicked(false);
    setTimeout(() => {
      setIsRotateRightClicked(false);
    }, 300);
    handleSeek(15);
  };

  const handlePlayPause = () => {
    const videoElement = videoRef.current;
    if (!videoElement) return;

    if (isPlaying) {
      videoElement.pause();
    } else {
      videoElement.play();
    }
    setIsPlaying(!isPlaying);
  };

  const handleLastVideoOpenData = useCallback(
    (videoTime: number) => {
      const fileMinute = `${String(Math.floor(videoTime / 60)).padStart(
        2,
        "0"
      )}:${String(Math.floor(videoTime % 60)).padStart(2, "0")}`;

      console.debug("updating video data...", fileMinute);
      axios
        .post(
          `http://192.168.3.150:8080/episodes/${episodeId}/last-access?time=${fileMinute}`
        )
        .then(() => {
          console.debug("video data updated...");
        });
    },
    [episodeId]
  );

  const formatTime = (seconds: number) => {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = Math.floor(seconds % 60);
    return `${minutes}:${remainingSeconds.toString().padStart(2, "0")}`;
  };

  const handleProgressBarClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (videoRef.current) {
      const rect = e.currentTarget.getBoundingClientRect();
      const clickX = e.clientX - rect.left;
      const progress = (clickX / rect.width) * 100;
      videoRef.current.currentTime =
        (progress / 100) * videoRef.current.duration;
      handleLastVideoOpenData(videoRef.current.currentTime);
    }
  };

  const handleSkipToWhereYouLeft = useCallback(() => {
    if (videoRef.current && leftAt !== "00:00") {
      const [minutes, seconds] = leftAt.split(":");
      const timeInSeconds = parseInt(minutes) * 60 + parseInt(seconds);
      videoRef.current.currentTime = timeInSeconds;
    }
  }, [leftAt]);

  const handleSeek = (seconds: number) => {
    const videoElement = videoRef.current;
    if (!videoElement) return;
    videoElement.currentTime += seconds;
  };

  const handleTimeUpdate = () => {
    if (videoRef.current) {
      setCurrentTime(videoRef.current.currentTime);
      setDuration(videoRef.current.duration);
    }
  };

  const handleMouseMove = () => {
    setShowControls(true);
    if (controlsTimeoutRef.current) {
      clearTimeout(controlsTimeoutRef.current);
    }
    controlsTimeoutRef.current = setTimeout(() => {
      setShowControls(false);
    }, 3000);
  };

  useEffect(() => {
    const videoElement = videoRef.current;

    if (!videoElement) return;

    setTimeout(() => {
      setIsPlaying(true);
      handleSkipToWhereYouLeft();
      videoElement.play();
    }, 2000);

    const timeShifter = setInterval(() => {
      handleLastVideoOpenData(videoElement.currentTime);
    }, 60000);

    const hideControls = () => {
      setShowControls(false);
    };

    controlsTimeoutRef.current = setTimeout(hideControls, 3000);

    videoElement.addEventListener("ended", () => {
      handleLastVideoOpenData(videoElement.currentTime);
      onEnded && onEnded();
    });

    return () => {
      if (controlsTimeoutRef.current) {
        clearTimeout(controlsTimeoutRef.current);
      }
      clearInterval(timeShifter);
      handleLastVideoOpenData(videoElement.currentTime);
      setTimeout(() => {
        onClose();
      }, 1000);
      videoElement.removeEventListener("ended", () => {
        handleLastVideoOpenData(videoElement.currentTime);
        onEnded && onEnded();
      });
    };
  }, [handleLastVideoOpenData, handleSkipToWhereYouLeft, onClose, onEnded]);

  const progress = (currentTime / duration) * 100;

  const handleDoubleClick = () => {
    if (videoRef.current) {
      videoRef.current.requestFullscreen();
    }
  };

  return (
    <>
      {isOpen && (
        <VideoContainer onMouseMove={handleMouseMove}>
          <VideoPlayer
            ref={videoRef}
            src={url}
            onEnded={onEnded}
            onTimeUpdate={handleTimeUpdate}
            onDoubleClick={handleDoubleClick}
          />
          <ControlsContainer showControls={showControls}>
            <ControlButton onClick={handleRotateLeftClick}>
              <LeftIcon
                className="fas fa-rotate-left"
                rotated={isRotateLeftClicked}
              />
            </ControlButton>
            <PauseButton onClick={handlePlayPause}>
              {isPlaying ? (
                <i className="fas fa-pause"></i>
              ) : (
                <i className="fas fa-play"></i>
              )}
            </PauseButton>
            <ControlButton onClick={handleRotateRightClick}>
              <RightIcon
                className="fas fa-rotate-right"
                rotated={isRotateRightClicked}
              />
            </ControlButton>
          </ControlsContainer>
          <TimeInfo showControls={showControls}>
            <span>{formatTime(currentTime)}</span>
            <span>{formatTime(duration)}</span>
          </TimeInfo>
          <ProgressBarContainer
            onClick={handleProgressBarClick}
            showControls={showControls}
          >
            <ProgressBar progress={progress}>
              <ProgressIndicator progress={progress} />
            </ProgressBar>
          </ProgressBarContainer>
        </VideoContainer>
      )}
    </>
  );
};

export default FullScreenVideo;
