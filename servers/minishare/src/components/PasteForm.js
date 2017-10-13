import React, { Component } from 'react';
import './PasteForm.css'

import { options } from './type_options';

class PasteForm extends Component {
  componentDidMount() {
    this.title.focus();
  }

  render() {
    const { type, title, body, onChange, onSubmit } = this.props;

    return (
      <div>
        <form id="paste-form" onSubmit={onSubmit}>
          <div id="paste-form-nav">
            <input type="text"
                   value={title}
                   id={"paste-form-title-input"}
                   placeholder="Title (optional)"
                   maxLength={100}
                   ref={(element) => { this.title = element }}
                   onChange={onChange.bind(this, "title")} />

            Type: <select onChange={onChange.bind(this, "type")}
                          value={type}>
              {options.map(opt => {
                  return (
                      <option key={"option-" + opt[0]} value={opt[1]}>
                        {opt[0]}
                      </option>
                  )
              })}
            </select>
          </div>
          <input id="paste-submit" type="submit" value="Encrypt and Save" />
          <br />
          <textarea id="paste-textarea"
                    value={body}
                    onChange={onChange.bind(this, "body")} />
        </form>
      </div>
    );
  }
}

export default PasteForm;
