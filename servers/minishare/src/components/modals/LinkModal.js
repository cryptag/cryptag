import React, { Component } from 'react';
import PropTypes from 'prop-types';

import { Modal, Button } from 'react-bootstrap';

import './LinkModal.css';

class LinkModal extends Component {
  constructor(props) {
    super(props);

    this.state = {
      copied: false
    }
  }

  copyURLToClipboard = (event) => {
    event.preventDefault();

    this.highlightURL();
    document.execCommand("copy");

    this.setState({
      copied: true
    })
  }

  componentDidMount() {
    this.highlightURL();
  }

  highlightURL = () => {
    this.input.focus();

    // TODO: Try to show beginning of URL, perhaps using something
    // similar to `this.input.setSelectionRange(0, 1000, 'backward')`
    // (but not quite)
    this.input.select();
  }

  onFocus = (event) => {
    event.target.select();
  }

  onCloseModal = () => {
    this.props.onCloseModal();

    this.setState({
      copied: false
    });
  }

  render() {
    const { showModal, url } = this.props;
    const { copied } = this.state;

    return (
      <div>
        <Modal show={showModal} onHide={this.onCloseModal}>
          <Modal.Header closeButton>
            <Modal.Title>The Download URL</Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <h5>
              The data you're sharing via this link will be destroyed in 24 hours <strong><em>or</em></strong> when this link is first used -- whichever comes first.
            </h5>
            <div className="form-group">
              <input type="text"
                     value={url}
                     ref={(element) => { this.input = element }}
                     onFocus={this.onFocus}
                     style={{width: "100%"}} />
            </div>
            <h6>
              (Happy sharing, and thanks for using miniShare!)
            </h6>
          </Modal.Body>
          <Modal.Footer>
            <Button bsStyle={!copied ? 'primary' : 'success'}
                    ref={(btn) => { this.copyBtn = btn } }
                    onClick={this.copyURLToClipboard}>
              {!copied && 'Copy URL to Clipboard'}
              {copied && 'Copied URL to Clipboard!'}
            </Button>
            <Button onClick={this.onCloseModal}>Close</Button>
          </Modal.Footer>
        </Modal>
      </div>
    );
  }
}


LinkModal.propType = {
  showModal: PropTypes.bool.isRequired,
  onCloseModal: PropTypes.func.isRequired,
}

export default LinkModal;
