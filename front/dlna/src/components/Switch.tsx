import React from "react";
import classes from "./Switch.module.css";

const Switch = ({
  onChange,
}: {
  onChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
}) => {
  return (
    <label className={classes.switch}>
      <input type="checkbox" onChange={onChange} />
      <span className={`${classes.slider} ${classes.round}`}></span>
    </label>
  );
};

export default Switch;
