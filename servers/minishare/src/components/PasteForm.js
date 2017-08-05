import React, { Component } from 'react';
import './PasteForm.css'

import { options } from './type_options';

class PasteForm extends Component {
  constructor() {
    super();

    this.state = {
      pasteType: options[0][1],
      pasteTitle: '',
      pasteBody: '',
    }
  }

  onChange = (fieldName, event) => {
    event.preventDefault();
    this.setState({
      [fieldName]: event.target.value
    })
  }

  render() {
    return (
      <div>
        <form className="paste-form" onSubmit={this.props.onSubmit.bind(this, this.state)} >
          <div id="paste-form-nav">
            <input type="text"
                   placeholder="Title (optional)"
                   onChange={this.onChange.bind(this, "pasteTitle")} />

            {" | "}

            Type: <select>
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
                    onChange={this.onChange.bind(this, "pasteBody")} />
        </form>
      </div>
    );
  }
}

export default PasteForm;
