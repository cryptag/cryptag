import React, { Component } from 'react';
import PropTypes from 'prop-types';

import { Modal, Button } from 'react-bootstrap';

class LinkModal extends Component {
  componentDidMount() {
    this.input.focus();

    // TODO: Try to show beginning of URL, perhaps using something
    // similar to `this.input.setSelectionRange(0, 1000, 'backward')`
    // (but not quite)
    this.input.select();
  }

  onFocus = (event) => {
    event.target.select();
  }

  render() {
    let { showModal, url, onCloseModal } = this.props;

    return (
      <div>
        <Modal show={showModal} onHide={onCloseModal}>
          <Modal.Header closeButton>
            <Modal.Title>The Download URL</Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <div className="form-group">
              <input type="text"
                     value={url}
                     ref={(element) => { this.input = element }}
                     onFocus={this.onFocus}
                     style={{width: "100%"}} />
            </div>
          </Modal.Body>
          <Modal.Footer>
            <Button onClick={onCloseModal}>Close</Button>
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
