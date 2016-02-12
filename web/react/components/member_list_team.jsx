// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

import MemberListTeamItem from './member_list_team_item.jsx';
import UserStore from '../stores/user_store.jsx';
import {isGuest} from '../utils/utils.jsx';

export default class MemberListTeam extends React.Component {
    constructor(props) {
        super(props);

        this.getUsers = this.getUsers.bind(this);
        this.onChange = this.onChange.bind(this);

        this.state = {
            users: this.getUsers()
        };
    }

    componentDidMount() {
        UserStore.addChangeListener(this.onChange);
    }

    componentWillUnmount() {
        UserStore.removeChangeListener(this.onChange);
    }

    getUsers() {
        const profiles = UserStore.getProfiles();
        let users = [];

        for (const id of Object.keys(profiles)) {
            users.push(profiles[id]);
        }

        users = users.filter((user) => {
            return !isGuest(user.roles);
        });
        users.sort((a, b) => a.username.localeCompare(b.username));

        return users;
    }

    onChange() {
        this.setState({
            users: this.getUsers()
        });
    }

    render() {
        const memberList = this.state.users.map((user) => {
            return (
                <MemberListTeamItem
                    key={user.id}
                    user={user}
                />
            );
        });

        return (
            <table className='table more-table member-list-holder'>
                <tbody>
                    {memberList}
                </tbody>
            </table>
        );
    }
}
