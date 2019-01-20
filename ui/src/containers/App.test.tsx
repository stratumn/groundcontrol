import { shallow } from "enzyme";
import React from "react";

import App from "./App";

describe("<App />", () => {
  it("renders without crashing", () => {
    shallow(<App />);
  });

  it("renders children when passed in", () => {
    const wrapper = shallow((
      <App>
        <div className="unique" />
      </App>
    ));
    expect(wrapper.contains(<div className="unique" />)).toEqual(true);
  });
});
