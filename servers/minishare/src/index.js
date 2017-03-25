var React = require('react');
var ReactDOM = require('react-dom');

export default class App extends React.Component {
  constructor(){
    super(...arguments);

    this.state = {
      username: '',
      showUsernameModal: false,
      alertMessage: 'Welcome to miniShare!',
      alertStyle: 'success'
    };

    this.promptForUsername = this.promptForUsername.bind(this);
    this.loadUsername = this.loadUsername.bind(this);
  }

  promptForUsername(){
    this.setState({
      showUsernameModal: true
    });
  }

  loadUsername(){
    let { username } = this.state;

    if (!username){
      this.promptForUsername();
    }
  }

  render(){
    let { username, showUsernameModal } = this.state;

    let hash = document.location.hash;

    return (
      <div className="app">
        <h2>miniShare</h2>
      </div>
    )
  }
}

App.propTypes = {}

ReactDOM.render(<App />, document.getElementById('root'));
