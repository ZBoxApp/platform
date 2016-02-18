/**
 * Created by enahum on 2/8/16.
 */

import ErrorBar from '../components/error_bar.jsx';
import AddonController from '../components/addons/addon_controller.jsx';
import InstallAddonModal from '../components/addons/install_addon_modal.jsx';
import UninstallAddonModal from '../components/addons/uninstall_addon_modal.jsx';
import ConfigAddonModal from '../components/addons/config_addon_modal.jsx';
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
                <div>
                    <ErrorBar/>
                    <AddonController
                        category={this.props.map.Category}
                    />
                    <InstallAddonModal />
                    <UninstallAddonModal />
                    <ConfigAddonModal />
                </div>
            </IntlProvider>
        );
    }
}

global.window.setup_addons_page = function setup(props) {
    ReactDOM.render(
        <Root map={props} />,
        document.getElementById('addons')
    );
};
