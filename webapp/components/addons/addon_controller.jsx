/**
 * Created by enahum on 2/8/16.
 */

import React from 'react';
import UserStore from 'stores/user_store.jsx';
import AddonStore from 'stores/addon_store.jsx';

import LoadingScreen from '../loading_screen.jsx';
import AddonSidebar from './addon_sidebar.jsx';
import AddonContainer from './addon_container.jsx';

import InstallAddonModal from './install_addon_modal.jsx';
import UninstallAddonModal from './uninstall_addon_modal.jsx';
import ConfigAddonModal from './config_addon_modal.jsx';

import ErrorPage from '../../components/error_page.jsx';

import * as Utils from 'utils/utils.jsx';
import * as Client from 'utils/client.jsx';

import {injectIntl, intlShape, defineMessages, FormattedMessage} from 'react-intl';

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

        this.state = {
            loaded: false
        };
    }
    componentDidMount() {
        UserStore.addChangeListener(this.getStateFromProps);
    }
    componentWillUnmount() {
        UserStore.removeChangeListener(this.getStateFromProps);
    }
    getStateFromProps() {
        const cs = this.loadCategories();
        const selected = cs[0].id;
        const me = UserStore.getCurrentUser();
        const isAdmin = Utils.isAdmin(me.roles);

        this.setState({
            loaded: true,
            isAdmin,
            categories: cs,
            selected
        });

        this.getAddons();
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
        var tab = <LoadingScreen/>;

        if (!this.state.loaded) {
            return tab;
        }

        if (!this.state.isAdmin) {
            return (
                <ErrorPage
                    message={
                        <FormattedMessage
                            id='addons.unauthorized'
                            defaultMessage="You're not Authorized to manage addons for this Team"
                        />
                    }
                    goBack={this.props.history.goBack}
                />
            );
        }

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
            <div id='addons'>
                <div>
                    <div
                        className='sidebar--menu'
                        id='sidebar-menu'
                    ></div>
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
                <InstallAddonModal/>
                <UninstallAddonModal/>
                <ConfigAddonModal/>
            </div>
        );
    }
}

AddonController.propTypes = {
    intl: intlShape.isRequired,
    history: React.PropTypes.object
};

export default injectIntl(AddonController);