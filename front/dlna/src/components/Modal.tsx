import React, { useEffect } from "react";
import classes from "./Modal.module.css";

const Modal = ({
  isOpen,
  onClose,
  children,
}: {
  isOpen: boolean;
  onClose: () => void;
  children: React.ReactNode;
}) => {
  useEffect(() => {
    const closeOnBackdropClick = (event: Event) => {
      if (
        isOpen &&
        (event.target as HTMLElement).classList.contains("backdrop")
      ) {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener("click", closeOnBackdropClick);
    }

    return () => {
      document.removeEventListener("click", closeOnBackdropClick);
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div className={classes.modal}>
      <div className={classes.backdrop} onClick={onClose}></div>
      <div className={classes.modal_content}>
        <span className={classes.close} onClick={onClose}>
          &times;
        </span>
        <div>{children}</div>
      </div>
    </div>
  );
};

export default Modal;
