/**
 * Created by enahum on 2/8/16.
 */
import AddonSidebarHeader from './addon_sidebar_header.jsx';

import {FormattedMessage} from 'mm-intl';

export default class AddonSidebar extends React.Component {
    constructor(props) {
        super(props);

        this.isSelected = this.isSelected.bind(this);
        this.handleClick = this.handleClick.bind(this);
    }

    componentDidMount() {
        if ($(window).width() > 768) {
            $('.nav-pills__container').perfectScrollbar();
        }
    }

    handleClick(category, e) {
        e.preventDefault();
        this.props.selectTab(category);
        history.pushState({category}, null, `/addons/${category}`);
    }

    isSelected(category) {
        if (this.props.selected === category) {
            return 'active';
        }

        return '';
    }

    render() {
        const categories = this.props.categories.map((c) => {
            return (
                <li key={c.id}>
                    <a
                        href='#'
                        className={this.isSelected(c.id)}
                        onClick={this.handleClick.bind(this, c.id)}
                    >
                        {c.name}
                    </a>
                </li>
            );
        });

        return (
            <div className='sidebar--left sidebar--collapsable'>
                <div>
                    <AddonSidebarHeader />
                    <div className='nav-pills__container'>
                        <ul className='nav nav-pills nav-stacked'>
                            <li>
                                <ul className='nav nav__sub-menu'>
                                    <li>
                                        <h4>
                                            <span className='icon fa fa-puzzle-piece'></span>
                                            <span>
                                                <FormattedMessage
                                                    id='addon.sidebar.categories'
                                                    defaultMessage='Categories'
                                                />
                                            </span>
                                        </h4>
                                    </li>
                                </ul>
                                <ul className='nav nav__sub-menu padded'>
                                    {categories}
                                </ul>
                            </li>
                        </ul>
                    </div>
                </div>
            </div>
        );
    }
}

AddonSidebar.propTypes = {
    categories: React.PropTypes.array.isRequired,
    selected: React.PropTypes.string.isRequired,
    selectTab: React.PropTypes.func.isRequired
};