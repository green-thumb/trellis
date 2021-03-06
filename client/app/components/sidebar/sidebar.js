var React = require('react');
var Chat = require('./sidebar-chat');
var ChatActions = require('../../actions/ChatActions');
var ChatStore = require('../../stores/ChatStore');
var AuthStore = require('../../stores/AuthStore');
var rd3 = require('react-d3');
var PieChart = rd3.PieChart;
var Treemap = rd3.Treemap;

var FriendList = require("../friend/friendlist")

// TODO - factor out navbar login form

var getTrending = function(callback) {
  $.ajax({
    type: 'GET',
    url: '/trending/',
    crossDomain: true,
    success: function(resp) { // WORKING for fetchuser?
      // console.log('success',resp);
      callback(resp);
    },
    error: function(resp) {
      // TODO: Fix this, this always goes to error - not sure.
      // Found out - jQuery 1.4.2 works with current go server, but breaks with newer ver.
      console.log('error',resp);
      callback(null);
    }
  });

};

var Sidebar = React.createClass({

    getInitialState: function(){
      return {
        from: "",
        messages: [],
        data : []
      };
    },

    loadTrending: function(){
      // Get Trending data
      var that = this;

      getTrending(function(data){
        var array = [];
        var total = 0;
        for (var i = 0; i < data.topics.length; i++) {
          total += data.topics[i].count;
        };
        for (var i = 0; i < data.topics.length; i++) {
          var obj = data.topics[i];
          array.push({label: obj.tag, value: Math.round((obj.count/total)*100)});
        };
        that.setState({
          data:array
        });
      });
    },

    componentDidMount: function(){
      AuthStore.addChangeListener(this._onAuthChange);
      ChatStore.addChangeListener(this._onChange);
      this.loadTrending();
    },

    componentWillUnmount: function(){
      AuthStore.removeChangeListener(this._onAuthChange);
      ChatStore.removeChangeListener(this._onChange);
    },

    _onChange: function(){
        this.setState({
          messages: ChatStore.getMessages()
        });  
    },

    _onAuthChange: function(){
      this.setState({
        from: AuthStore.getUser().username
      });
      this.joinChat(); // On Auth change, if user logs in then connect to chat server.
    },

    joinChat: function(){
      ChatActions.connect();
    },

    sendMessage: function(msg){
      ChatActions.send({message:msg});
    },

    render: function(){

    return (
        <ul className="sidebar-nav">
            <a href="#"><img src="/assets/Trellis-logo.png" ></img></a>
            <li>
                <a href="#">Trending (past 24 hours)</a>
                <Treemap
                  data={this.state.data}
                  width={225}
                  height={200}
                  textColor="#484848"
                  fontSize="12px"/>
            </li>
            <li>
                <a href="#">Chat (global)</a>
            </li>
            <Chat messages={this.state.messages} user={this.state.from} onSend={this.sendMessage} onChat={this.joinChat} />
            <div className="friend-online-list"> Friends Online </div>
            <FriendList />
        </ul>
    );
  }
});

module.exports = Sidebar;
