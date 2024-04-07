import React from "react";
import "./ModeSwitch.css";

interface ModeSwitchProps {
  onChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
}

const ModeSwitch: React.FC<ModeSwitchProps> = ({ onChange }) => {
  return (
    <div className="switch">
      <input
        type="checkbox"
        className="switch__input"
        onChange={onChange}
        id="Switch"
      />
      <label className="switch__label" htmlFor="Switch">
        <span className="switch__indicator"></span>
        <span className="switch__decoration"></span>
      </label>
    </div>
  );
};

export default ModeSwitch;
