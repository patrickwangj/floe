import {Panel} from '../panel/panel.js';
import {RestCall} from '../panel/rest.js';

"use strict";

// the controller for the Dashboard
export function Header() {
    var panel = {};

    function evtLogout() {
        RestCall(panel.evtHub, "POST", "/logout");
    }

    var events = [
        {El: '#logout', Ev: 'click', Fn: evtLogout}
    ];

    panel = new Panel(this, {}, tpl, 'header', events);

    // check if we have a session cookie
    console.log(document.cookie);
    var i = document.cookie.indexOf("floe-sesh=");
    if (i >= 0) {
        console.log("got floe sesh");
        panel.store.Update("Authed", true); // assume token is valid - time will tell
    }

    this.Map = function(evt) {
        var data = {};
        if (evt.Type == 'unauth') {
            data.Authed = false;
        }
        if (evt.Type == 'auth') {
            data.Authed = true;
        }
        // TODO map the event data to the panel data model
        return data;
    }

    return panel;
}

var tpl = `
<h3 class='title'><a href='/dash'>Dash</a> > Build BE</h3>
<nav>
    <ul>
        <li><a href='/settings'>Settings</a></li>
        {{? it.Data.Authed }}
        <li><a id='logout' href='/logout'>Logout</a></li>
        {{?}}
        {{? !it.Data.Authed }}
        <li><a href='/login'>Login</a></li>
        {{?}}
    </ul>
</nav>
`