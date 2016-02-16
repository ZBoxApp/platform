// Copyright (c) 2016 ZBox, Spa. All Rights Reserved.
// See License.txt for license information.

import SignupUserComplete from '../components/guest_signup.jsx';
import * as Client from '../utils/client.jsx';

var IntlProvider = ReactIntl.IntlProvider;

class Root extends React.Component {
    constructor() {
        super();
        this.state = {
            translations: null,
            loaded: false
        };
    }

    static propTypes() {
        return {
            map: React.PropTypes.object.isRequired
        };
    }

    componentWillMount() {
        Client.getTranslations(
            this.props.map.Locale,
            (data) => {
                this.setState({
                    translations: data,
                    loaded: true
                });
            },
            () => {
                this.setState({
                    loaded: true
                });
            }
        );
    }

    render() {
        if (!this.state.loaded) {
            return <div></div>;
        }

        return (
            <IntlProvider
                locale={this.props.map.Locale}
                messages={this.state.translations}
            >
                <SignupUserComplete
                    teamId={this.props.map.TeamId}
                    teamName={this.props.map.TeamName}
                    teamDisplayName={this.props.map.TeamDisplayName}
                    email={this.props.map.Email}
                    hash={this.props.map.Hash}
                    data={this.props.map.Data}
                />
            </IntlProvider>
        );
    }
}

global.window.setup_guest_signup_page = function setup(props) {
    ReactDOM.render(
        <Root map={props} />,
        document.getElementById('guest-signup')
    );
};