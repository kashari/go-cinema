import React from "react";
import { Navbar, Nav, Container } from "react-bootstrap";
import gopher from "../assets/gopher.png";
import { NavLink } from "react-router-dom";
import ModeSwitch from "./ModeSwitch";
import "../App.css";

const NavigationPanel: React.FC = () => {
  return (
    <Navbar bg="transparent" expand="lg" className="colorized">
      <Container>
        <Navbar.Brand style={{ fontSize: "30px", fontWeight: "bold" }}>
          <img src={gopher} style={{ height: "100px" }} />
        </Navbar.Brand>
        <Navbar.Toggle aria-controls="basic-navbar-nav" />
        <Navbar.Collapse id="basic-navbar-nav">
          <Nav className="me-auto">
            <Nav.Item
              style={{
                marginRight: "26px",
                marginTop: "16px",
                fontSize: "50px",
              }}
            >
              <NavLink
                to="/"
                style={{ textDecoration: "none", color: "#89CFF0" }}
              >
                Home
              </NavLink>
            </Nav.Item>
            <Nav.Item
              style={{
                marginRight: "26px",
                marginTop: "16px",
                fontSize: "50px",
              }}
            >
              <NavLink
                to="/movies"
                style={{ textDecoration: "none", color: "#89CFF0" }}
              >
                Movies
              </NavLink>
            </Nav.Item>
            <Nav.Item
              style={{
                marginRight: "26px",
                marginTop: "16px",
                fontSize: "50px",
              }}
            >
              <NavLink
                to="/series"
                style={{ textDecoration: "none", color: "#89CFF0" }}
              >
                Series
              </NavLink>
            </Nav.Item>

            <Nav.Item
              style={{
                marginRight: "26px",
                marginTop: "16px",
                fontSize: "50px",
              }}
            >
              <NavLink
                to="/management"
                style={{ textDecoration: "none", color: "#89CFF0" }}
              >
                Management
              </NavLink>
            </Nav.Item>
          </Nav>
          <Nav.Item
            style={{
              marginRight: "16px",
              marginLeft: "16px",
              marginTop: "26px",
            }}
          >
            <ModeSwitch
              onChange={() => {
                if (
                  document.documentElement.getAttribute("data-bs-theme") ==
                  "dark"
                ) {
                  document.documentElement.setAttribute(
                    "data-bs-theme",
                    "light"
                  );
                } else {
                  document.documentElement.setAttribute(
                    "data-bs-theme",
                    "dark"
                  );
                }
              }}
            />
          </Nav.Item>
        </Navbar.Collapse>
      </Container>
    </Navbar>
  );
};

export default NavigationPanel;
