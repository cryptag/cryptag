import React, { Component } from 'react';
import './PasteForm.css'

import { options } from './type_options';

class PasteForm extends Component {
  render() {
    return (
      <div>
        <form className="paste-form"
              onSubmit={this.props.onSubmit} >

          <div id="paste-form-nav">
            <input type="text"
                   placeholder="Title (optional)"
                   onChange={this.props.onChange.bind(this, "pasteTitle")} />

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
                    onChange={this.props.onChange.bind(this, "pasteBody")} />
        </form>
      </div>
    );
  }
}

export default PasteForm;
