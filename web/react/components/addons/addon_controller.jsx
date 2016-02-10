/**
 * Created by enahum on 2/8/16.
 */

import LoadingScreen from '../loading_screen.jsx';
import AddonSidebar from './addon_sidebar.jsx';
import AddonContainer from './addon_container.jsx';
import AddonStore from '../../stores/addon_store.jsx';
import * as Utils from '../../utils/utils.jsx';
import * as Client from '../../utils/client.jsx';

import {injectIntl, intlShape, defineMessages} from 'mm-intl';

const categories = defineMessages({
    analytics: {
        id: 'addons.category.analytics',
        defaultMessage: 'Analytics'
    },
    bots: {
        id: 'addons.category.bots',
        defaultMessage: 'Awesome Bots'
    },
    communication: {
        id: 'addons.category.communication',
        defaultMessage: 'Communication'
    },
    customer: {
        id: 'addons.category.support',
        defaultMessage: 'Customer Support'
    },
    design: {
        id: 'addons.category.design',
        defaultMessage: 'Design'
    },
    developer: {
        id: 'addons.category.developer',
        defaultMessage: 'Developer Tools'
    },
    entretainment: {
        id: 'addons.category.entretainment',
        defaultMessage: 'Entretainment'
    },
    file: {
        id: 'addons.category.file',
        defaultMessage: 'File Management'
    },
    hr: {
        id: 'addons.category.hr',
        defaultMessage: 'HR'
    },
    marketing: {
        id: 'addons.category.marketing',
        defaultMessage: 'Marketing'
    },
    office: {
        id: 'addons.category.office',
        defaultMessage: 'Office Management'
    },
    payments: {
        id: 'addons.category.payments',
        defaultMessage: 'Payments & Accounting'
    },
    productivity: {
        id: 'addons.category.productivity',
        defaultMessage: 'Productivity'
    },
    project: {
        id: 'addons.category.project',
        defaultMessage: 'Project Management'
    },
    security: {
        id: 'addons.category.security',
        defaultMessage: 'Security'
    },
    social: {
        id: 'addons.category.social',
        defaultMessage: 'Social'
    }
});

class AddonController extends React.Component {
    constructor(props) {
        super(props);

        this.loadCategories = this.loadCategories.bind(this);
        this.selectTab = this.selectTab.bind(this);
        this.getStateFromProps = this.getStateFromProps.bind(this);
        this.getAddons = this.getAddons.bind(this);

        this.state = this.getStateFromProps();
        this.getAddons();

        if (!props.category) {
            history.replaceState(null, null, `/addons/${this.state.selected}`);
        }
    }

    getStateFromProps() {
        const cs = this.loadCategories();
        const selected = this.props.category || cs[0].id;

        return {
            categories: cs,
            selected: selected
        };
    }

    getAddons() {
        const state = this.state;

        Client.listAddons(
            (data) => {
                if (data) {
                    AddonStore.storeAddons(data);
                }

                state.getAddonsComplete = true;
                this.setState(state);
            },
            (err) => {
                this.setState({serverError: err});
            }
        );
    }

    loadCategories() {
        const {formatMessage} = this.props.intl;

        const array = Object.keys(categories).map((key) => {
            return {
                id: key,
                name: formatMessage(categories[key])
            };
        });

        return Utils.sortByKey(array, 'name');
    }

    selectTab(category) {
        const state = this.state;
        state.getAddonsComplete = true;
        state.selected = category;
        this.setState(state);
    }

    render() {
        var tab = <LoadingScreen />;

        if (this.state.getAddonsComplete) {
            tab = (
                <AddonContainer
                    error={this.state.serverError}
                    addons={AddonStore.getAddons(this.state.selected)}
                    category={this.state.selected}
                />
            );
        }

        return (
            <div>
                <div
                    className='sidebar--menu'
                    id='sidebar-menu'
                />
                <AddonSidebar
                    categories={this.state.categories}
                    selected={this.state.selected}
                    selectTab={this.selectTab}
                />
                <div className='inner__wrap channel__wrap'>
                    <div className='row header'>
                    </div>
                    <div className='row main'>
                        <div
                            id='app-content'
                            className='app__content addon'
                        >
                            {tab}
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}

AddonController.propTypes = {
    intl: intlShape.isRequired,
    category: React.PropTypes.string
};

export default injectIntl(AddonController);