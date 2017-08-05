import React, { Component } from 'react';
import './App.css';

import PasteForm from './PasteForm';

class App extends Component {
  constructor() {
    super();

    this.state = {
    }
  }

  onPasteSubmit = ({pasteType, pasteTitle, pasteBody}, event) => {
    event.preventDefault();
    console.log(`TODO: encrypt and send pasteTitle '${pasteTitle}' with pasteBody '${pasteBody}', not to mention pasteType ${pasteType}`);
  }

  render() {
    return (
      <div className="App">
        <div className="App-header">
          <h2>miniShare</h2>
        </div>

        <h4>Securely share self-destructing data: text, URLs, and (soon!) files</h4>

        <PasteForm
          onSubmit={this.onPasteSubmit} />

        <div id="footer">
          <a href="https://www.leapchat.org/">LeapChat</a>: End-to-end encrypted in-browser chat!
          {" | "}
          <a href="https://www.patreon.com/cryptag">Support me on Patreon!</a>
        </div>
      </div>
    );
  }
}

export default App;
