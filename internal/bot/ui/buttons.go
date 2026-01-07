package ui

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Locale string

const (
	LocaleRU Locale = "ru"
	LocaleEN Locale = "en"
)

const DefaultLocale = LocaleRU

const (
	IconSSL        = "🔒"
	IconNoSSL      = "🔓"
	IconChart      = "📊"
	IconArrowLeft  = "◀️"
	IconArrowRight = "▶️"
)

type ButtonID string

const (
	MainMenu ButtonID = "main_menu"
	Cancel   ButtonID = "cancel"
	Back     ButtonID = "back"

	AddClient                  ButtonID = "add_client"
	RepeatCheck                ButtonID = "repeat_check"
	NewSearch                  ButtonID = "new_search"
	PrevPage                   ButtonID = "prev_page"
	NextPage                   ButtonID = "next_page"
	Yes                        ButtonID = "yes"
	No                         ButtonID = "no"
	CheckTorrents              ButtonID = "check_torrents"
	QuickActionsMenu           ButtonID = "quick_actions_menu"
	ClientsMenu                ButtonID = "clients_menu"
	AddTorrentFile             ButtonID = "add_torrent_file"
	SearchTorrent              ButtonID = "search_torrent"
	MonitorTorrent             ButtonID = "monitor_torrent"
	BackToTorrents             ButtonID = "back_to_torrents"
	SubscribeNotifyBot         ButtonID = "subscribe_notify_bot"
	PauseTorrentsMenu          ButtonID = "pause_torrents_menu"
	ResumeTorrentsMenu         ButtonID = "resume_torrents_menu"
	PauseAllTorrents           ButtonID = "pause_all_torrents"
	ResumeAllTorrents          ButtonID = "resume_all_torrents"
	PauseRutrackerTorrents     ButtonID = "pause_rutracker_torrents"
	ResumeRutrackerTorrents    ButtonID = "resume_rutracker_torrents"
	PauseNonRutrackerTorrents  ButtonID = "pause_non_rutracker_torrents"
	ResumeNonRutrackerTorrents ButtonID = "resume_non_rutracker_torrents"
	SpeedLimitMenu             ButtonID = "speed_limit_menu"
	SettingsMenu               ButtonID = "settings_menu"
	Variables                  ButtonID = "variables"
	RecommendedTorrents        ButtonID = "recommended_torrents"
	Timezone                   ButtonID = "timezone"

	Speed10       ButtonID = "speed_10"
	Speed100      ButtonID = "speed_100"
	Speed500      ButtonID = "speed_500"
	Speed1000     ButtonID = "speed_1000"
	Speed2000     ButtonID = "speed_2000"
	Speed5000     ButtonID = "speed_5000"
	Speed10000    ButtonID = "speed_10000"
	Speed50000    ButtonID = "speed_50000"
	SpeedCustom   ButtonID = "speed_custom"
	SpeedRemove   ButtonID = "speed_remove"
	Edit          ButtonID = "edit"
	Delete        ButtonID = "delete"
	BackToList    ButtonID = "back_to_list"
	ConfirmDelete ButtonID = "confirm_delete"
	KeepExisting  ButtonID = "keep_existing"

	ManualHashInput ButtonID = "manual_hash_input"
	ManualPathInput ButtonID = "manual_path_input"
	SkipHashYes     ButtonID = "skip_hash_yes"
	SkipHashNo      ButtonID = "skip_hash_no"
	DeleteFilesYes  ButtonID = "delete_files_yes"
	DeleteFilesNo   ButtonID = "delete_files_no"
	PauseTorrent    ButtonID = "pause_torrent"
	ResumeTorrent   ButtonID = "resume_torrent"
)

type MsgID string

const (
	MsgUnknownCommand                        MsgID = "unknown_command"
	MsgClientsListError                      MsgID = "clients_list_error"
	MsgClientsListErrorWithEmoji             MsgID = "clients_list_error_with_emoji"
	MsgCheckAllClientsListError              MsgID = "check_all_clients_list_error"
	MsgCheckAllNoClients                     MsgID = "check_all_no_clients"
	MsgCheckAllCheckingNClients              MsgID = "check_all_checking_n_clients"
	MsgCheckAllCheckingClientsProgress       MsgID = "check_all_checking_clients_progress"
	MsgCheckAllSingleClientChecking          MsgID = "check_all_single_client_checking"
	MsgCheckAllSingleClientConnectOKGetting  MsgID = "check_all_single_client_connect_ok_getting"
	MsgCheckAllSingleClientConnectError      MsgID = "check_all_single_client_connect_error"
	MsgCheckAllSingleClientGetTorrentsError  MsgID = "check_all_single_client_get_torrents_error"
	MsgCheckAllSingleClientFiltering         MsgID = "check_all_single_client_filtering"
	MsgCheckAllSingleClientFilterError       MsgID = "check_all_single_client_filter_error"
	MsgCheckAllSingleClientCheckingRutracker MsgID = "check_all_single_client_checking_rutracker"
	MsgCheckAllSingleClientRutrackerAPIError MsgID = "check_all_single_client_rutracker_api_error"
	MsgCheckAllResultsStale                  MsgID = "check_all_results_stale"
	MsgCheckAllResultsHeader                 MsgID = "check_all_results_header"
	MsgCheckAllResultsTotalTimeFmt           MsgID = "check_all_results_total_time_fmt"
	MsgCheckAllResultsLastCheckFmt           MsgID = "check_all_results_last_check_fmt"
	MsgCheckAllResultsSeparator              MsgID = "check_all_results_separator"
	MsgCheckAllResultsClientErrorFmt         MsgID = "check_all_results_client_error_fmt"
	MsgCheckAllResultsClientLineFmt          MsgID = "check_all_results_client_line_fmt"
	MsgCheckAllResultsActiveFmt              MsgID = "check_all_results_active_fmt"
	MsgCheckAllResultsFilteredFmt            MsgID = "check_all_results_filtered_fmt"
	MsgCheckAllResultsActualFmt              MsgID = "check_all_results_actual_fmt"
	MsgCheckAllResultsMissingCountFmt        MsgID = "check_all_results_missing_count_fmt"
	MsgCheckAllResultsMissingShownFirstFmt   MsgID = "check_all_results_missing_shown_first_fmt"
	MsgCheckAllResultsMissingItemFmt         MsgID = "check_all_results_missing_item_fmt"
	MsgCheckAllResultsDurationFmt            MsgID = "check_all_results_duration_fmt"
	MsgErrorInvalidClientID                  MsgID = "error_invalid_client_id"
	MsgErrorInvalidTorrentIndex              MsgID = "error_invalid_torrent_index"
	MsgErrorInvalidResultIndex               MsgID = "error_invalid_result_index"
	MsgErrorInvalidSpeedValue                MsgID = "error_invalid_speed_value"
	MsgErrorDataNotFoundStartOver            MsgID = "error_data_not_found_start_over"
	MsgErrorTorrentDataNotFoundStartOver     MsgID = "error_torrent_data_not_found_start_over"
	MsgErrorPathNotFound                     MsgID = "error_path_not_found"
	MsgErrorInvalidPath                      MsgID = "error_invalid_path"
	MsgErrorSavePathNotSelectedStartOver     MsgID = "error_save_path_not_selected_start_over"
	MsgErrorGetClientData                    MsgID = "error_get_client_data"
	MsgErrorGetClientDataWithEmoji           MsgID = "error_get_client_data_with_emoji"
	MsgErrorClientNotFoundOrNoAccess         MsgID = "error_client_not_found_or_no_access"
	MsgErrorDeleteClientTryAgain             MsgID = "error_delete_client_try_again"
	MsgErrorConnectClientFmt                 MsgID = "error_connect_client_fmt"
	MsgErrorPauseTorrent                     MsgID = "error_pause_torrent"
	MsgErrorResumeTorrent                    MsgID = "error_resume_torrent"
	MsgErrorAddTorrentFmt                    MsgID = "error_add_torrent_fmt"
	MsgErrorDeleteTorrentFmt                 MsgID = "error_delete_torrent_fmt"
	MsgErrorSendTorrentFilePrompt            MsgID = "error_send_torrent_file_prompt"
	MsgAddClientCancelled                    MsgID = "add_client_cancelled"
	MsgEditClientCancelled                   MsgID = "edit_client_cancelled"
	MsgVariablesRecommendedTorrentsPrompt    MsgID = "variables_recommended_torrents_prompt"
	MsgQuickActionsNoClients                 MsgID = "quick_actions_no_clients"
	MsgDialogUnknownStateStartOver           MsgID = "dialog_unknown_state_start_over"
	MsgDialogInvalidStateAddStartOver        MsgID = "dialog_invalid_state_add_start_over"
	MsgDialogInvalidStateEditStartOver       MsgID = "dialog_invalid_state_edit_start_over"
	MsgDialogInvalidStateStartOver           MsgID = "dialog_invalid_state_start_over"
	MsgDialogSessionExpiredStartOver         MsgID = "dialog_session_expired_start_over"
	MsgDialogInvalidDataStartOver            MsgID = "dialog_invalid_data_start_over"
	MsgDialogInvalidPortStartOver            MsgID = "dialog_invalid_port_start_over"
	MsgDialogInvalidClientIDStartOver        MsgID = "dialog_invalid_client_id_start_over"
	MsgDialogInvalidClientID                 MsgID = "dialog_invalid_client_id"
	MsgDialogPortMustBeNumberTryAgain        MsgID = "dialog_port_must_be_number_try_again"
	MsgDialogAddClientStart                  MsgID = "dialog_add_client_start"
	MsgDialogEnterClientName                 MsgID = "dialog_enter_client_name"
	MsgDialogEnterHost                       MsgID = "dialog_enter_host"
	MsgDialogEnterPort                       MsgID = "dialog_enter_port"
	MsgDialogEnterUsername                   MsgID = "dialog_enter_username"
	MsgDialogEnterPassword                   MsgID = "dialog_enter_password"
	MsgDialogUseSSL                          MsgID = "dialog_use_ssl"
	MsgDialogCreateClientErrorTryAgain       MsgID = "dialog_create_client_error_try_again"
	MsgDialogEditClientStartFmt              MsgID = "dialog_edit_client_start_fmt"
	MsgDialogUpdateClientErrorTryAgain       MsgID = "dialog_update_client_error_try_again"
	MsgDialogRecommendedEmptyPrompt          MsgID = "dialog_recommended_empty_prompt"
	MsgDialogRecommendedInvalidPrompt        MsgID = "dialog_recommended_invalid_prompt"
	MsgDialogRecommendedSaveErrorTryAgain    MsgID = "dialog_recommended_save_error_try_again"
	MsgDocumentProcessingTorrentFileFmt      MsgID = "document_processing_torrent_file_fmt"
	MsgClientsForTorrentAddNoClients         MsgID = "clients_for_torrent_add_no_clients"
	MsgClientsForTorrentAddChooseClient      MsgID = "clients_for_torrent_add_choose_client"
	MsgClientsForTorrentMonitorNoClients     MsgID = "clients_for_torrent_monitor_no_clients"
	MsgClientsForTorrentMonitorChooseClient  MsgID = "clients_for_torrent_monitor_choose_client"
	MsgClientDetailsSSLYES                   MsgID = "client_details_ssl_yes"
	MsgClientDetailsSSLNO                    MsgID = "client_details_ssl_no"
	MsgClientDetailsFmt                      MsgID = "client_details_fmt"
	MsgDeleteClientConfirmationFmt           MsgID = "delete_client_confirmation_fmt"
	MsgCheckClientsListNoClients             MsgID = "check_clients_list_no_clients"
	MsgCheckClientsListHeader                MsgID = "check_clients_list_header"
	MsgCheckClientsListChooseClient          MsgID = "check_clients_list_choose_client"
	MsgCheckClientsListButtonCheckFmt        MsgID = "check_clients_list_button_check_fmt"
	MsgClientsListNoClients                  MsgID = "clients_list_no_clients"
	MsgClientsListHeader                     MsgID = "clients_list_header"
	MsgClientsListButtonDetailsFmt           MsgID = "clients_list_button_details_fmt"
	MsgDialogMonitorHashEmptyPrompt          MsgID = "dialog_monitor_hash_empty_prompt"
	MsgDialogMonitorHashLengthPrompt         MsgID = "dialog_monitor_hash_length_prompt"
	MsgDialogSearchQueryEmptyPrompt          MsgID = "dialog_search_query_empty_prompt"
	MsgDialogSpeedEmptyPrompt                MsgID = "dialog_speed_empty_prompt"
	MsgDialogSpeedInvalidFormatPrompt        MsgID = "dialog_speed_invalid_format_prompt"
	MsgDialogSpeedMustBePositivePrompt       MsgID = "dialog_speed_must_be_positive_prompt"
	MsgSearchNoResultsFmt                    MsgID = "search_no_results_fmt"
	MsgSearchResultsHeaderWithQueryFmt       MsgID = "search_results_header_with_query_fmt"
	MsgSearchResultsHeader                   MsgID = "search_results_header"
	MsgSearchResultsFoundCountFmt            MsgID = "search_results_found_count_fmt"
	MsgSearchResultsItemLineFmt              MsgID = "search_results_item_line_fmt"
	MsgSearchResultsItemHashFmt              MsgID = "search_results_item_hash_fmt"
	MsgCheckClientNoActiveFmt                MsgID = "check_client_no_active_fmt"
	MsgCheckClientResultHeaderFmt            MsgID = "check_client_result_header_fmt"
	MsgCheckClientMissingCountFmt            MsgID = "check_client_missing_count_fmt"
	MsgCheckClientMissingShownFirstFmt       MsgID = "check_client_missing_shown_first_fmt"
	MsgCheckClientMissingItemFmt             MsgID = "check_client_missing_item_fmt"
	MsgCheckClientDurationFooterFmt          MsgID = "check_client_duration_footer_fmt"

	MsgMainMenuClientConnectErrorFmt      MsgID = "main_menu_client_connect_error_fmt"
	MsgMainMenuClientTransferInfoErrorFmt MsgID = "main_menu_client_transfer_info_error_fmt"
	MsgMainMenuClientLinePrefixFmt        MsgID = "main_menu_client_line_prefix_fmt"
	MsgMainMenuDownloadFmt                MsgID = "main_menu_download_fmt"
	MsgMainMenuUploadFmt                  MsgID = "main_menu_upload_fmt"
	MsgMainMenuSpeedLimitBracketFmt       MsgID = "main_menu_speed_limit_bracket_fmt"
	MsgMainMenuNoClientsText              MsgID = "main_menu_no_clients_text"

	MsgSpeedZero         MsgID = "speed_zero"
	MsgSpeedBpsFmt       MsgID = "speed_bps_fmt"
	MsgSpeedKBpsFmt      MsgID = "speed_kbps_fmt"
	MsgSpeedMBpsFmt      MsgID = "speed_mbps_fmt"
	MsgSpeedLimitMBpsFmt MsgID = "speed_limit_mbps_fmt"

	MsgClientListItemFmt    MsgID = "client_list_item_fmt"
	MsgClientHostPortFmt    MsgID = "client_host_port_fmt"
	MsgClientButtonLabelFmt MsgID = "client_button_label_fmt"

	MsgSettingsMenuText             MsgID = "settings_menu_text"
	MsgNotificationsStatusOn        MsgID = "notifications_status_on"
	MsgNotificationsStatusOff       MsgID = "notifications_status_off"
	MsgNotificationsToggleButtonFmt MsgID = "notifications_toggle_button_fmt"

	MsgTimezoneMenuText               MsgID = "timezone_menu_text"
	MsgVariablesMenuText              MsgID = "variables_menu_text"
	MsgKeyValueIntFmt                 MsgID = "key_value_int_fmt"
	MsgKeyValueStringFmt              MsgID = "key_value_string_fmt"
	MsgEditRecommendedTorrentsTextFmt MsgID = "edit_recommended_torrents_text_fmt"

	MsgSearchStartPromptText      MsgID = "search_start_prompt_text"
	MsgSearchClientsListError     MsgID = "search_clients_list_error"
	MsgSearchNoClientsText        MsgID = "search_no_clients_text"
	MsgSearchResultsStaleText     MsgID = "search_results_stale_text"
	MsgSearchResultsButtonItemFmt MsgID = "search_results_button_item_fmt"

	MsgQuickActionsMenuText          MsgID = "quick_actions_menu_text"
	MsgQuickActionsMenuNoClientsText MsgID = "quick_actions_menu_no_clients_text"
	MsgQuickActionsPauseMenuText     MsgID = "quick_actions_pause_menu_text"
	MsgQuickActionsResumeMenuText    MsgID = "quick_actions_resume_menu_text"

	MsgSpeedLimitMenuText          MsgID = "speed_limit_menu_text"
	MsgSpeedLimitMenuNoClientsText MsgID = "speed_limit_menu_no_clients_text"
	MsgSpeedLimitCustomPromptText  MsgID = "speed_limit_custom_prompt_text"
	MsgSpeedLimitAppliedHeaderFmt  MsgID = "speed_limit_applied_header_fmt"
	MsgSpeedLimitsRemovedHeader    MsgID = "speed_limits_removed_header"

	MsgPauseAllHeaderText                 MsgID = "pause_all_header_text"
	MsgResumeAllHeaderText                MsgID = "resume_all_header_text"
	MsgPauseAllClientSuccessFmt           MsgID = "pause_all_client_success_fmt"
	MsgResumeAllClientSuccessFmt          MsgID = "resume_all_client_success_fmt"
	MsgPauseRutrackerHeaderText           MsgID = "pause_rutracker_header_text"
	MsgResumeRutrackerHeaderText          MsgID = "resume_rutracker_header_text"
	MsgPauseRutrackerClientSuccessFmt     MsgID = "pause_rutracker_client_success_fmt"
	MsgResumeRutrackerClientSuccessFmt    MsgID = "resume_rutracker_client_success_fmt"
	MsgPauseNonRutrackerHeaderText        MsgID = "pause_non_rutracker_header_text"
	MsgResumeNonRutrackerHeaderText       MsgID = "resume_non_rutracker_header_text"
	MsgPauseNonRutrackerClientSuccessFmt  MsgID = "pause_non_rutracker_client_success_fmt"
	MsgResumeNonRutrackerClientSuccessFmt MsgID = "resume_non_rutracker_client_success_fmt"
	MsgResultErrorsHeaderFmt              MsgID = "result_errors_header_fmt"
	MsgResultErrorsItemFmt                MsgID = "result_errors_item_fmt"
	MsgResultTotalsFmt                    MsgID = "result_totals_fmt"

	MsgBotCommandMenuDescription    MsgID = "bot_command_menu_description"
	MsgBotCommandCheckDescription   MsgID = "bot_command_check_description"
	MsgBotCommandClientsDescription MsgID = "bot_command_clients_description"

	MsgTorrentMonitorSelectText       MsgID = "torrent_monitor_select_text"
	MsgTorrentMonitorManualHashPrompt MsgID = "torrent_monitor_manual_hash_prompt"
	MsgTorrentMonitorItemButtonFmt    MsgID = "torrent_monitor_item_button_fmt"

	MsgSavePathSelectionHeaderText              MsgID = "save_path_selection_header_text"
	MsgSavePathSelectionTorrentFmt              MsgID = "save_path_selection_torrent_fmt"
	MsgSavePathSelectionRecommendedBlockFmt     MsgID = "save_path_selection_recommended_block_fmt"
	MsgSavePathSelectionRecommendedButtonFmt    MsgID = "save_path_selection_recommended_button_fmt"
	MsgSavePathSelectionDefaultBlockFmt         MsgID = "save_path_selection_default_block_fmt"
	MsgSavePathSelectionDefaultButtonFmt        MsgID = "save_path_selection_default_button_fmt"
	MsgSavePathSelectionExistingPathsHeaderText MsgID = "save_path_selection_existing_paths_header_text"
	MsgSavePathSelectionPathButtonFmt           MsgID = "save_path_selection_path_button_fmt"
	MsgCustomSavePathPromptText                 MsgID = "custom_save_path_prompt_text"
	MsgSkipHashCheckQuestionFmt                 MsgID = "skip_hash_check_question_fmt"
	MsgDeleteExistingTorrentQuestionFmt         MsgID = "delete_existing_torrent_question_fmt"
	MsgDeleteFilesQuestionText                  MsgID = "delete_files_question_text"
	MsgTorrentDeletedFilesSuffix                MsgID = "torrent_deleted_files_suffix"
	MsgTorrentDeletedSuccessFmt                 MsgID = "torrent_deleted_success_fmt"
	MsgAddTorrentSendFilePromptFmt              MsgID = "add_torrent_send_file_prompt_fmt"

	MsgTorrentProgressHeaderText        MsgID = "torrent_progress_header_text"
	MsgTorrentProgressNameFmt           MsgID = "torrent_progress_name_fmt"
	MsgTorrentProgressPathFmt           MsgID = "torrent_progress_path_fmt"
	MsgTorrentProgressStatusFmt         MsgID = "torrent_progress_status_fmt"
	MsgTorrentProgressPercentFmt        MsgID = "torrent_progress_percent_fmt"
	MsgTorrentProgressDownloadFmt       MsgID = "torrent_progress_download_fmt"
	MsgTorrentProgressUploadFmt         MsgID = "torrent_progress_upload_fmt"
	MsgTorrentProgressUploadedFmt       MsgID = "torrent_progress_uploaded_fmt"
	MsgTorrentProgressSeedsPeersFmt     MsgID = "torrent_progress_seeds_peers_fmt"
	MsgTorrentProgressSizeFmt           MsgID = "torrent_progress_size_fmt"
	MsgTorrentProgressSpeedSuffixPerSec MsgID = "torrent_progress_speed_suffix_per_sec"

	MsgTorrentStatusCompleted    MsgID = "torrent_status_completed"
	MsgTorrentStatusPaused       MsgID = "torrent_status_paused"
	MsgTorrentStatusDownloading  MsgID = "torrent_status_downloading"
	MsgTorrentStatusError        MsgID = "torrent_status_error"
	MsgTorrentStatusMissingFiles MsgID = "torrent_status_missing_files"
	MsgTorrentStatusOtherFmt     MsgID = "torrent_status_other_fmt"

	MsgTorrentMonitorClientProcessingFmt        MsgID = "torrent_monitor_client_processing_fmt"
	MsgTorrentMonitorClientTorrentProcessingFmt MsgID = "torrent_monitor_client_torrent_processing_fmt"
	MsgTorrentMonitorOpenTorrentButtonText      MsgID = "torrent_monitor_open_torrent_button_text"

	MsgTorrentActivityDownloading      MsgID = "torrent_activity_downloading"
	MsgTorrentActivityUploading        MsgID = "torrent_activity_uploading"
	MsgTorrentActivityUploadStalled    MsgID = "torrent_activity_upload_stalled"
	MsgTorrentActivityDownloadStalled  MsgID = "torrent_activity_download_stalled"
	MsgTorrentActivityChecking         MsgID = "torrent_activity_checking"
	MsgTorrentActivityQueued           MsgID = "torrent_activity_queued"
	MsgTorrentActivityPaused           MsgID = "torrent_activity_paused"
	MsgTorrentActivityFetchingMetadata MsgID = "torrent_activity_fetching_metadata"
	MsgTorrentActivityError            MsgID = "torrent_activity_error"
	MsgTorrentActivityMissingFiles     MsgID = "torrent_activity_missing_files"
	MsgTorrentActivityOtherFmt         MsgID = "torrent_activity_other_fmt"
)

type buttonResource struct {
	text map[Locale]string
	data string
}

var buttonResources = map[ButtonID]buttonResource{
	MainMenu: {
		data: "main_menu",
		text: map[Locale]string{
			LocaleRU: "🏠 В главное меню",
			LocaleEN: "🏠 Main menu",
		},
	},
	Cancel: {
		text: map[Locale]string{
			LocaleRU: "❌ Отмена",
			LocaleEN: "❌ Cancel",
		},
	},
	Back: {
		text: map[Locale]string{
			LocaleRU: "🔙 Назад",
			LocaleEN: "🔙 Back",
		},
	},
	AddClient: {
		data: "add_client",
		text: map[Locale]string{
			LocaleRU: "➕ Добавить клиента",
			LocaleEN: "➕ Add client",
		},
	},
	RepeatCheck: {
		data: "check_torrents",
		text: map[Locale]string{
			LocaleRU: "🔄 Повторить проверку",
			LocaleEN: "🔄 Re-run check",
		},
	},
	NewSearch: {
		data: "search_torrent",
		text: map[Locale]string{
			LocaleRU: "🔎 Новый поиск",
			LocaleEN: "🔎 New search",
		},
	},
	PrevPage: {
		text: map[Locale]string{
			LocaleRU: "◀️ Назад",
			LocaleEN: "◀️ Back",
		},
	},
	NextPage: {
		text: map[Locale]string{
			LocaleRU: "Вперёд ▶️",
			LocaleEN: "Next ▶️",
		},
	},
	Yes: {
		text: map[Locale]string{
			LocaleRU: "✅ Да",
			LocaleEN: "✅ Yes",
		},
	},
	No: {
		text: map[Locale]string{
			LocaleRU: "❌ Нет",
			LocaleEN: "❌ No",
		},
	},
	CheckTorrents: {
		data: "check_torrents",
		text: map[Locale]string{
			LocaleRU: "🔍 Проверка обновлений",
			LocaleEN: "🔍 Check updates",
		},
	},
	QuickActionsMenu: {
		data: "quick_actions",
		text: map[Locale]string{
			LocaleRU: "⚡ Быстрые действия",
			LocaleEN: "⚡ Quick actions",
		},
	},
	ClientsMenu: {
		data: "clients",
		text: map[Locale]string{
			LocaleRU: "📋 Клиенты",
			LocaleEN: "📋 Clients",
		},
	},
	AddTorrentFile: {
		data: "add_torrent_file",
		text: map[Locale]string{
			LocaleRU: "📥 Добавить торрент файл",
			LocaleEN: "📥 Add torrent file",
		},
	},
	SearchTorrent: {
		data: "search_torrent",
		text: map[Locale]string{
			LocaleRU: "🔎 Поиск торрента",
			LocaleEN: "🔎 Search torrent",
		},
	},
	MonitorTorrent: {
		data: "monitor_torrent",
		text: map[Locale]string{
			LocaleRU: "📊 Мониторинг торрента",
			LocaleEN: "📊 Torrent monitoring",
		},
	},
	BackToTorrents: {
		data: "back_to_torrents",
		text: map[Locale]string{
			LocaleRU: "🔙 К списку торрентов",
			LocaleEN: "🔙 Back to torrents",
		},
	},
	SubscribeNotifyBot: {
		text: map[Locale]string{
			LocaleRU: "🔔 Подключить уведомления",
			LocaleEN: "🔔 Enable notifications",
		},
	},
	PauseTorrentsMenu: {
		data: "quick_action_pause_menu",
		text: map[Locale]string{
			LocaleRU: "⏸ Остановить торренты -> ...",
			LocaleEN: "⏸ Pause torrents -> ...",
		},
	},
	ResumeTorrentsMenu: {
		data: "quick_action_resume_menu",
		text: map[Locale]string{
			LocaleRU: "▶ Возобновить торренты -> ...",
			LocaleEN: "▶ Resume torrents -> ...",
		},
	},
	PauseAllTorrents: {
		data: "quick_action_pause_all",
		text: map[Locale]string{
			LocaleRU: "⏸ Остановить все раздачи",
			LocaleEN: "⏸ Pause all torrents",
		},
	},
	ResumeAllTorrents: {
		data: "quick_action_resume_all",
		text: map[Locale]string{
			LocaleRU: "▶ Запустить все раздачи",
			LocaleEN: "▶ Resume all torrents",
		},
	},
	PauseRutrackerTorrents: {
		data: "quick_action_pause_rutracker",
		text: map[Locale]string{
			LocaleRU: "⏸ Остановить rutracker раздачи",
			LocaleEN: "⏸ Pause rutracker torrents",
		},
	},
	ResumeRutrackerTorrents: {
		data: "quick_action_resume_rutracker",
		text: map[Locale]string{
			LocaleRU: "▶ Запустить rutracker раздачи",
			LocaleEN: "▶ Resume rutracker torrents",
		},
	},
	PauseNonRutrackerTorrents: {
		data: "quick_action_pause_non_rutracker",
		text: map[Locale]string{
			LocaleRU: "⏸ Остановить НЕ rutracker раздачи",
			LocaleEN: "⏸ Pause non-rutracker torrents",
		},
	},
	ResumeNonRutrackerTorrents: {
		data: "quick_action_resume_non_rutracker",
		text: map[Locale]string{
			LocaleRU: "▶ Запустить НЕ rutracker раздачи",
			LocaleEN: "▶ Resume non-rutracker torrents",
		},
	},
	SpeedLimitMenu: {
		data: "quick_action_limit_speed_menu",
		text: map[Locale]string{
			LocaleRU: "🚦 Ограничение скорости",
			LocaleEN: "🚦 Speed limit",
		},
	},
	Speed10:             {data: "quick_action_limit_speed_10", text: map[Locale]string{LocaleRU: "0.10 МБ/с", LocaleEN: "0.10 MB/s"}},
	Speed100:            {data: "quick_action_limit_speed_100", text: map[Locale]string{LocaleRU: "1.00 МБ/с", LocaleEN: "1.00 MB/s"}},
	Speed500:            {data: "quick_action_limit_speed_500", text: map[Locale]string{LocaleRU: "5.00 МБ/с", LocaleEN: "5.00 MB/s"}},
	Speed1000:           {data: "quick_action_limit_speed_1000", text: map[Locale]string{LocaleRU: "10.00 МБ/с", LocaleEN: "10.00 MB/s"}},
	Speed2000:           {data: "quick_action_limit_speed_2000", text: map[Locale]string{LocaleRU: "20.00 МБ/с", LocaleEN: "20.00 MB/s"}},
	Speed5000:           {data: "quick_action_limit_speed_5000", text: map[Locale]string{LocaleRU: "50.00 МБ/с", LocaleEN: "50.00 MB/s"}},
	Speed10000:          {data: "quick_action_limit_speed_10000", text: map[Locale]string{LocaleRU: "100.00 МБ/с", LocaleEN: "100.00 MB/s"}},
	Speed50000:          {data: "quick_action_limit_speed_50000", text: map[Locale]string{LocaleRU: "500.00 МБ/с", LocaleEN: "500.00 MB/s"}},
	SpeedCustom:         {data: "quick_action_limit_speed_custom", text: map[Locale]string{LocaleRU: "✏ Ввести вручную", LocaleEN: "✏ Enter manually"}},
	SpeedRemove:         {data: "quick_action_remove_speed_limits", text: map[Locale]string{LocaleRU: "🚫 Убрать все ограничения", LocaleEN: "🚫 Remove limits"}},
	Edit:                {text: map[Locale]string{LocaleRU: "✏️ Изменить", LocaleEN: "✏️ Edit"}},
	Delete:              {text: map[Locale]string{LocaleRU: "🗑 Удалить", LocaleEN: "🗑 Delete"}},
	BackToList:          {data: "clients", text: map[Locale]string{LocaleRU: "🔙 Назад к списку", LocaleEN: "🔙 Back to list"}},
	ConfirmDelete:       {text: map[Locale]string{LocaleRU: "✅ Да, удалить", LocaleEN: "✅ Yes, delete"}},
	KeepExisting:        {text: map[Locale]string{LocaleRU: "❌ Нет, оставить", LocaleEN: "❌ No, keep"}},
	ManualHashInput:     {text: map[Locale]string{LocaleRU: "✏️ Ввести хеш вручную", LocaleEN: "✏️ Enter hash manually"}},
	ManualPathInput:     {text: map[Locale]string{LocaleRU: "✏️ Ввести путь вручную", LocaleEN: "✏️ Enter path manually"}},
	SettingsMenu:        {data: "settings", text: map[Locale]string{LocaleRU: "⚙️ Настройки", LocaleEN: "⚙️ Settings"}},
	Timezone:            {data: "edit_timezone", text: map[Locale]string{LocaleRU: "🕒 Часовой пояс", LocaleEN: "🕒 Timezone"}},
	Variables:           {data: "variables", text: map[Locale]string{LocaleRU: "🔧 Переменные", LocaleEN: "🔧 Variables"}},
	RecommendedTorrents: {data: "edit_recommended_torrents", text: map[Locale]string{LocaleRU: "📌 Рекомендуемое количество торрентов", LocaleEN: "📌 Recommended torrents count"}},
	SkipHashYes:         {text: map[Locale]string{LocaleRU: "✅ Да, пропустить", LocaleEN: "✅ Yes, skip"}},
	SkipHashNo:          {text: map[Locale]string{LocaleRU: "❌ Нет, проверить", LocaleEN: "❌ No, check"}},
	DeleteFilesYes:      {text: map[Locale]string{LocaleRU: "✅ Да, удалить файлы", LocaleEN: "✅ Yes, delete files"}},
	DeleteFilesNo:       {text: map[Locale]string{LocaleRU: "❌ Нет, только торрент", LocaleEN: "❌ No, torrent only"}},
	PauseTorrent: {
		data: "monitor_pause",
		text: map[Locale]string{
			LocaleRU: "⏸ Остановить",
			LocaleEN: "⏸ Pause torrent",
		},
	},
	ResumeTorrent: {
		data: "monitor_resume",
		text: map[Locale]string{
			LocaleRU: "▶ Запустить раздачу",
			LocaleEN: "▶ Resume torrent",
		},
	},
}

var msgResources = map[MsgID]map[Locale]string{
	MsgUnknownCommand: {
		LocaleRU: "Неизвестная команда",
		LocaleEN: "Unknown command",
	},
	MsgClientsListError: {
		LocaleRU: "Ошибка при получении списка клиентов",
		LocaleEN: "Failed to load clients list",
	},
	MsgClientsListErrorWithEmoji: {
		LocaleRU: "❌ Ошибка при получении списка клиентов",
		LocaleEN: "❌ Failed to load clients list",
	},
	MsgCheckAllClientsListError: {
		LocaleRU: "❌ Ошибка при получении списка клиентов",
		LocaleEN: "❌ Failed to load clients list",
	},
	MsgCheckAllNoClients: {
		LocaleRU: "📋 *Проверка активных торрентов*\n\nКлиенты не найдены. Добавьте клиента для проверки.",
		LocaleEN: "📋 *Active torrents check*\n\nNo clients found. Add a client to run the check.",
	},
	MsgCheckAllCheckingNClients: {
		LocaleRU: "🔍 Проверка активных торрентов для *%d* клиентов...",
		LocaleEN: "🔍 Checking active torrents for *%d* clients...",
	},
	MsgCheckAllCheckingClientsProgress: {
		LocaleRU: "🔍 Проверка клиентов...\n\n*%d* из *%d*\n\nПроверка: *%s*",
		LocaleEN: "🔍 Checking clients...\n\n*%d* of *%d*\n\nChecking: *%s*",
	},
	MsgCheckAllSingleClientChecking: {
		LocaleRU: "🔍 Проверка активных торрентов для клиента *%s*...",
		LocaleEN: "🔍 Checking active torrents for client *%s*...",
	},
	MsgCheckAllSingleClientConnectOKGetting: {
		LocaleRU: "✅ Подключение к *%s* успешно\n\n🔍 Получение списка активных торрентов...",
		LocaleEN: "✅ Connected to *%s*\n\n🔍 Getting active torrents...",
	},
	MsgCheckAllSingleClientConnectError: {
		LocaleRU: "❌ Ошибка при подключении к клиенту *%s*:\n`%v`",
		LocaleEN: "❌ Failed to connect to client *%s*:\n`%v`",
	},
	MsgCheckAllSingleClientGetTorrentsError: {
		LocaleRU: "❌ Ошибка при получении торрентов от клиента *%s*:\n`%v`",
		LocaleEN: "❌ Failed to get torrents from client *%s*:\n`%v`",
	},
	MsgCheckAllSingleClientFiltering: {
		LocaleRU: "✅ Подключение к *%s* успешно\n\n🔍 Получено активных торрентов: *%d*\n\n🔍 Фильтрация по комментарию (rutracker)...",
		LocaleEN: "✅ Connected to *%s*\n\n🔍 Active torrents: *%d*\n\n🔍 Filtering by comment (rutracker)...",
	},
	MsgCheckAllSingleClientFilterError: {
		LocaleRU: "❌ Ошибка при фильтрации торрентов от клиента *%s*:\n`%v`",
		LocaleEN: "❌ Failed to filter torrents from client *%s*:\n`%v`",
	},
	MsgCheckAllSingleClientCheckingRutracker: {
		LocaleRU: "✅ Подключение к *%s* успешно\n\n🔍 Получено активных торрентов: *%d*\n\n✅ Отфильтровано по rutracker: *%d*\n\n🔍 Проверка хешей в API рутрекера...",
		LocaleEN: "✅ Connected to *%s*\n\n🔍 Active torrents: *%d*\n\n✅ Filtered (rutracker): *%d*\n\n🔍 Checking hashes in rutracker API...",
	},
	MsgCheckAllSingleClientRutrackerAPIError: {
		LocaleRU: "❌ Ошибка при проверке хешей в API рутрекера от клиента *%s*:\n`%v`",
		LocaleEN: "❌ Failed to check hashes in rutracker API for client *%s*:\n`%v`",
	},
	MsgCheckAllResultsStale: {
		LocaleRU: "Результаты проверки устарели. Запустите проверку заново.",
		LocaleEN: "Check results are stale. Run the check again.",
	},
	MsgCheckAllResultsHeader: {
		LocaleRU: "📊 *Результаты проверки всех клиентов*\n\n",
		LocaleEN: "📊 *All clients check results*\n\n",
	},
	MsgCheckAllResultsTotalTimeFmt: {
		LocaleRU: "⏱ Общее время: *%s*\n",
		LocaleEN: "⏱ Total time: *%s*\n",
	},
	MsgCheckAllResultsLastCheckFmt: {
		LocaleRU: "🕐 Последняя проверка: *%s*\n",
		LocaleEN: "🕐 Last check: *%s*\n",
	},
	MsgCheckAllResultsSeparator: {
		LocaleRU: "\n---\n\n",
		LocaleEN: "\n---\n\n",
	},
	MsgCheckAllResultsClientErrorFmt: {
		LocaleRU: "❌ *%s*\n   `%s`\n\n",
		LocaleEN: "❌ *%s*\n   `%s`\n\n",
	},
	MsgCheckAllResultsClientLineFmt: {
		LocaleRU: "💻 *%s*\n",
		LocaleEN: "💻 *%s*\n",
	},
	MsgCheckAllResultsActiveFmt: {
		LocaleRU: "   📊 Активных: *%d*\n",
		LocaleEN: "   📊 Active: *%d*\n",
	},
	MsgCheckAllResultsFilteredFmt: {
		LocaleRU: "   🔍 Отфильтровано: *%d*\n",
		LocaleEN: "   🔍 Filtered: *%d*\n",
	},
	MsgCheckAllResultsActualFmt: {
		LocaleRU: "   ✅ Актуальных: *%d*/*%d*\n",
		LocaleEN: "   ✅ Actual: *%d*/*%d*\n",
	},
	MsgCheckAllResultsMissingCountFmt: {
		LocaleRU: "   ⚠️ Не найдено: *%d*\n\n",
		LocaleEN: "   ⚠️ Missing: *%d*\n\n",
	},
	MsgCheckAllResultsMissingShownFirstFmt: {
		LocaleRU: "   _Показано первых %d из %d:_\n\n",
		LocaleEN: "   _Showing first %d of %d:_\n\n",
	},
	MsgCheckAllResultsMissingItemFmt: {
		LocaleRU: "   • `%s`\n     `%s`\n",
		LocaleEN: "   • `%s`\n     `%s`\n",
	},
	MsgCheckAllResultsDurationFmt: {
		LocaleRU: "   ⏱ *%s*\n\n",
		LocaleEN: "   ⏱ *%s*\n\n",
	},
	MsgErrorInvalidClientID: {
		LocaleRU: "Ошибка: неверный ID клиента",
		LocaleEN: "Error: invalid client ID",
	},
	MsgErrorInvalidTorrentIndex: {
		LocaleRU: "❌ Ошибка: неверный индекс торрента.",
		LocaleEN: "❌ Error: invalid torrent index.",
	},
	MsgErrorInvalidResultIndex: {
		LocaleRU: "Ошибка: неверный индекс результата",
		LocaleEN: "Error: invalid result index",
	},
	MsgErrorInvalidSpeedValue: {
		LocaleRU: "Ошибка: неверное значение скорости",
		LocaleEN: "Error: invalid speed value",
	},
	MsgErrorDataNotFoundStartOver: {
		LocaleRU: "❌ Ошибка: данные не найдены. Начните заново.",
		LocaleEN: "❌ Error: data not found. Start over.",
	},
	MsgErrorTorrentDataNotFoundStartOver: {
		LocaleRU: "❌ Ошибка: данные торрента не найдены. Начните заново.",
		LocaleEN: "❌ Error: torrent data not found. Start over.",
	},
	MsgErrorPathNotFound: {
		LocaleRU: "❌ Ошибка: путь не найден",
		LocaleEN: "❌ Error: path not found",
	},
	MsgErrorInvalidPath: {
		LocaleRU: "❌ Ошибка: неверный путь",
		LocaleEN: "❌ Error: invalid path",
	},
	MsgErrorSavePathNotSelectedStartOver: {
		LocaleRU: "❌ Ошибка: путь сохранения не выбран. Начните заново.",
		LocaleEN: "❌ Error: save path not selected. Start over.",
	},
	MsgErrorGetClientData: {
		LocaleRU: "Ошибка при получении данных клиента",
		LocaleEN: "Failed to load client data",
	},
	MsgErrorGetClientDataWithEmoji: {
		LocaleRU: "❌ Ошибка при получении данных клиента",
		LocaleEN: "❌ Failed to load client data",
	},
	MsgErrorClientNotFoundOrNoAccess: {
		LocaleRU: "Клиент не найден или у вас нет доступа",
		LocaleEN: "Client not found or you don't have access",
	},
	MsgErrorDeleteClientTryAgain: {
		LocaleRU: "Ошибка при удалении клиента. Попробуйте снова.",
		LocaleEN: "Failed to delete client. Try again.",
	},
	MsgErrorConnectClientFmt: {
		LocaleRU: "❌ Ошибка при подключении к клиенту *%s*",
		LocaleEN: "❌ Failed to connect to client *%s*",
	},
	MsgErrorPauseTorrent: {
		LocaleRU: "❌ Ошибка при остановке торрента",
		LocaleEN: "❌ Failed to pause torrent",
	},
	MsgErrorResumeTorrent: {
		LocaleRU: "❌ Ошибка при запуске торрента",
		LocaleEN: "❌ Failed to resume torrent",
	},
	MsgErrorAddTorrentFmt: {
		LocaleRU: "❌ Ошибка при добавлении торрента: %v",
		LocaleEN: "❌ Failed to add torrent: %v",
	},
	MsgErrorDeleteTorrentFmt: {
		LocaleRU: "❌ Ошибка при удалении торрента: %v",
		LocaleEN: "❌ Failed to delete torrent: %v",
	},
	MsgErrorSendTorrentFilePrompt: {
		LocaleRU: "❌ Пожалуйста, отправьте файл с расширением .torrent",
		LocaleEN: "❌ Please send a file with .torrent extension",
	},
	MsgAddClientCancelled: {
		LocaleRU: "Добавление клиента отменено",
		LocaleEN: "Client adding cancelled",
	},
	MsgEditClientCancelled: {
		LocaleRU: "Редактирование клиента отменено",
		LocaleEN: "Client editing cancelled",
	},
	MsgVariablesRecommendedTorrentsPrompt: {
		LocaleRU: "✏️ Введите число рекомендуемых торрентов для отображения на странице выбора мониторинга (например: 3):",
		LocaleEN: "✏️ Enter recommended torrents count to display on monitoring selection page (e.g. 3):",
	},
	MsgQuickActionsNoClients: {
		LocaleRU: "❌ Клиенты не найдены",
		LocaleEN: "❌ No clients found",
	},
	MsgDialogUnknownStateStartOver: {
		LocaleRU: "Ошибка: неизвестное состояние. Начните операцию заново.",
		LocaleEN: "Error: unknown state. Start over.",
	},
	MsgDialogInvalidStateAddStartOver: {
		LocaleRU: "Ошибка: неверное состояние. Начните добавление заново.",
		LocaleEN: "Error: invalid state. Start adding again.",
	},
	MsgDialogInvalidStateEditStartOver: {
		LocaleRU: "Ошибка: неверное состояние. Начните редактирование заново.",
		LocaleEN: "Error: invalid state. Start editing again.",
	},
	MsgDialogInvalidStateStartOver: {
		LocaleRU: "Ошибка: неверное состояние. Начните заново.",
		LocaleEN: "Error: invalid state. Start over.",
	},
	MsgDialogSessionExpiredStartOver: {
		LocaleRU: "Ошибка: сессия истекла. Начните заново.",
		LocaleEN: "Error: session expired. Start over.",
	},
	MsgDialogInvalidDataStartOver: {
		LocaleRU: "Ошибка: неверные данные. Начните заново.",
		LocaleEN: "Error: invalid data. Start over.",
	},
	MsgDialogInvalidPortStartOver: {
		LocaleRU: "Ошибка: неверный порт. Начните заново.",
		LocaleEN: "Error: invalid port. Start over.",
	},
	MsgDialogInvalidClientIDStartOver: {
		LocaleRU: "Ошибка: неверный ID клиента. Начните заново.",
		LocaleEN: "Error: invalid client ID. Start over.",
	},
	MsgDialogInvalidClientID: {
		LocaleRU: "Ошибка: неверный ID клиента.",
		LocaleEN: "Error: invalid client ID.",
	},
	MsgDialogPortMustBeNumberTryAgain: {
		LocaleRU: "⚠️ Ошибка: порт должен быть числом. Попробуйте снова:",
		LocaleEN: "⚠️ Error: port must be a number. Try again:",
	},
	MsgDialogAddClientStart: {
		LocaleRU: "➕ *Добавление нового клиента*\n\n📝 Введите имя клиента:",
		LocaleEN: "➕ *Add new client*\n\n📝 Enter client name:",
	},
	MsgDialogEnterClientName: {
		LocaleRU: "📝 Введите имя клиента:",
		LocaleEN: "📝 Enter client name:",
	},
	MsgDialogEnterHost: {
		LocaleRU: "🌐 Введите host (например: 192.168.1.100):",
		LocaleEN: "🌐 Enter host (e.g. 192.168.1.100):",
	},
	MsgDialogEnterPort: {
		LocaleRU: "🔌 Введите port (например: 8080):",
		LocaleEN: "🔌 Enter port (e.g. 8080):",
	},
	MsgDialogEnterUsername: {
		LocaleRU: "👤 Введите username:",
		LocaleEN: "👤 Enter username:",
	},
	MsgDialogEnterPassword: {
		LocaleRU: "🔑 Введите password:",
		LocaleEN: "🔑 Enter password:",
	},
	MsgDialogUseSSL: {
		LocaleRU: "🔒 Использовать SSL?",
		LocaleEN: "🔒 Use SSL?",
	},
	MsgDialogCreateClientErrorTryAgain: {
		LocaleRU: "Ошибка при создании клиента. Попробуйте снова.",
		LocaleEN: "Failed to create client. Try again.",
	},
	MsgDialogEditClientStartFmt: {
		LocaleRU: "✏️ *Редактирование клиента*\n\n📝 Текущее имя: `%s`\n\nВведите новое имя клиента (или отправьте текущее для сохранения):",
		LocaleEN: "✏️ *Edit client*\n\n📝 Current name: `%s`\n\nEnter a new name (or send the current one to keep it):",
	},
	MsgDialogUpdateClientErrorTryAgain: {
		LocaleRU: "Ошибка при обновлении клиента. Попробуйте снова.",
		LocaleEN: "Failed to update client. Try again.",
	},
	MsgDialogRecommendedEmptyPrompt: {
		LocaleRU: "❌ Значение не может быть пустым. Введите число (например: 3):",
		LocaleEN: "❌ Value can't be empty. Enter a number (e.g. 3):",
	},
	MsgDialogRecommendedInvalidPrompt: {
		LocaleRU: "❌ Неверный формат или значение. Введите целое число от 1 до 100:",
		LocaleEN: "❌ Invalid format or value. Enter an integer from 1 to 100:",
	},
	MsgDialogRecommendedSaveErrorTryAgain: {
		LocaleRU: "❌ Ошибка при сохранении настройки. Попробуйте снова.",
		LocaleEN: "❌ Failed to save setting. Try again.",
	},
	MsgDocumentProcessingTorrentFileFmt: {
		LocaleRU: "📥 *Обработка торрент файла*\n\n📎 Файл: `%s`\n\n⏳ Обрабатываю...",
		LocaleEN: "📥 *Processing torrent file*\n\n📎 File: `%s`\n\n⏳ Processing...",
	},
	MsgClientsForTorrentAddNoClients: {
		LocaleRU: "📥 *Добавление торрент файла*\n\nКлиенты не найдены. Добавьте клиента для загрузки торрента.",
		LocaleEN: "📥 *Add torrent file*\n\nNo clients found. Add a client to upload a torrent.",
	},
	MsgClientsForTorrentAddChooseClient: {
		LocaleRU: "📥 *Добавление торрент файла*\n\nВыберите клиент для загрузки торрента:",
		LocaleEN: "📥 *Add torrent file*\n\nChoose a client to upload a torrent:",
	},
	MsgClientsForTorrentMonitorNoClients: {
		LocaleRU: "📊 *Мониторинг торрента*\n\nКлиенты не найдены. Добавьте клиента для мониторинга.",
		LocaleEN: "📊 *Torrent monitoring*\n\nNo clients found. Add a client to start monitoring.",
	},
	MsgClientsForTorrentMonitorChooseClient: {
		LocaleRU: "📊 *Мониторинг торрента*\n\nВыберите клиент:",
		LocaleEN: "📊 *Torrent monitoring*\n\nChoose a client:",
	},
	MsgClientDetailsSSLYES: {
		LocaleRU: "Да",
		LocaleEN: "Yes",
	},
	MsgClientDetailsSSLNO: {
		LocaleRU: "Нет",
		LocaleEN: "No",
	},
	MsgClientDetailsFmt: {
		LocaleRU: "🔧 *%s*\n\nHost: `%s`\nPort: `%d`\nUsername: `%s`\nSSL: `%s`\n",
		LocaleEN: "🔧 *%s*\n\nHost: `%s`\nPort: `%d`\nUsername: `%s`\nSSL: `%s`\n",
	},
	MsgDeleteClientConfirmationFmt: {
		LocaleRU: "⚠️ *Подтверждение удаления*\n\nВы уверены, что хотите удалить клиента *%s*?\n\nЭто действие нельзя отменить!",
		LocaleEN: "⚠️ *Confirm deletion*\n\nAre you sure you want to delete client *%s*?\n\nThis action can't be undone!",
	},
	MsgCheckClientsListNoClients: {
		LocaleRU: "📋 *Проверка активных торрентов*\n\nКлиенты не найдены. Добавьте клиента для проверки.",
		LocaleEN: "📋 *Active torrents check*\n\nNo clients found. Add a client to run the check.",
	},
	MsgCheckClientsListHeader: {
		LocaleRU: "📋 *Проверка активных торрентов*\n\n",
		LocaleEN: "📋 *Active torrents check*\n\n",
	},
	MsgCheckClientsListChooseClient: {
		LocaleRU: "Выберите клиента для проверки:\n\n",
		LocaleEN: "Choose a client to check:\n\n",
	},
	MsgCheckClientsListButtonCheckFmt: {
		LocaleRU: "🔍 Проверить %s",
		LocaleEN: "🔍 Check %s",
	},
	MsgClientsListNoClients: {
		LocaleRU: "📋 *Клиенты qBittorrent*\n\nКлиенты не найдены. Добавьте первого клиента.",
		LocaleEN: "📋 *qBittorrent clients*\n\nNo clients found. Add your first client.",
	},
	MsgClientsListHeader: {
		LocaleRU: "📋 *Клиенты qBittorrent*\n\n",
		LocaleEN: "📋 *qBittorrent clients*\n\n",
	},
	MsgClientsListButtonDetailsFmt: {
		LocaleRU: "🔧 %s",
		LocaleEN: "🔧 %s",
	},
	MsgDialogMonitorHashEmptyPrompt: {
		LocaleRU: "❌ Хеш не может быть пустым. Введите хеш торрента:",
		LocaleEN: "❌ Hash can't be empty. Enter torrent hash:",
	},
	MsgDialogMonitorHashLengthPrompt: {
		LocaleRU: "❌ Хеш должен содержать 40 символов. Введите правильный хеш:",
		LocaleEN: "❌ Hash must be 40 characters long. Enter a valid hash:",
	},
	MsgDialogSearchQueryEmptyPrompt: {
		LocaleRU: "❌ Поисковый запрос не может быть пустым. Введите хеш или название торрента:",
		LocaleEN: "❌ Search query can't be empty. Enter hash or torrent name:",
	},
	MsgDialogSpeedEmptyPrompt: {
		LocaleRU: "❌ Скорость не может быть пустой. Введите скорость в МБ/с:",
		LocaleEN: "❌ Speed can't be empty. Enter speed in MB/s:",
	},
	MsgDialogSpeedInvalidFormatPrompt: {
		LocaleRU: "❌ Неверный формат скорости. Введите число (например: 2.5):",
		LocaleEN: "❌ Invalid speed format. Enter a number (e.g. 2.5):",
	},
	MsgDialogSpeedMustBePositivePrompt: {
		LocaleRU: "❌ Скорость должна быть больше 0. Введите корректное значение:",
		LocaleEN: "❌ Speed must be greater than 0. Enter a valid value:",
	},
	MsgSearchNoResultsFmt: {
		LocaleRU: "🔎 *Поиск торрента*\n\n❌ По запросу `%s` ничего не найдено",
		LocaleEN: "🔎 *Torrent search*\n\n❌ Nothing found for `%s`",
	},
	MsgSearchResultsHeaderWithQueryFmt: {
		LocaleRU: "🔎 *Результаты поиска*\n\nЗапрос: `%s`\n\n",
		LocaleEN: "🔎 *Search results*\n\nQuery: `%s`\n\n",
	},
	MsgSearchResultsHeader: {
		LocaleRU: "🔎 *Результаты поиска*\n\n",
		LocaleEN: "🔎 *Search results*\n\n",
	},
	MsgSearchResultsFoundCountFmt: {
		LocaleRU: "Найдено: *%d*\n\n",
		LocaleEN: "Found: *%d*\n\n",
	},
	MsgSearchResultsItemLineFmt: {
		LocaleRU: "%d. *%s*\n",
		LocaleEN: "%d. *%s*\n",
	},
	MsgSearchResultsItemHashFmt: {
		LocaleRU: "   Hash: `%s`\n\n",
		LocaleEN: "   Hash: `%s`\n\n",
	},
	MsgCheckClientNoActiveFmt: {
		LocaleRU: "✅ *%s*\n\n📊 Активных торрентов: *%d*\n\n⏱ Время выполнения: *%s*\n\nНет активных торрентов",
		LocaleEN: "✅ *%s*\n\n📊 Active torrents: *%d*\n\n⏱ Duration: *%s*\n\nNo active torrents",
	},
	MsgCheckClientResultHeaderFmt: {
		LocaleRU: "💻 *%s*\n\n📊 Активных торрентов: *%d*\n\n🔍 Отфильтровано по rutracker: *%d*\n\n✅ Актуальных: *%d*/*%d*\n\n",
		LocaleEN: "💻 *%s*\n\n📊 Active torrents: *%d*\n\n🔍 Filtered by rutracker: *%d*\n\n✅ Actual: *%d*/*%d*\n\n",
	},
	MsgCheckClientMissingCountFmt: {
		LocaleRU: "\n\n⚠️ Не найдено в рутрекере: *%d*\n\n",
		LocaleEN: "\n\n⚠️ Missing on rutracker: *%d*\n\n",
	},
	MsgCheckClientMissingShownFirstFmt: {
		LocaleRU: "_Показано первых %d из %d:_\n\n",
		LocaleEN: "_Showing first %d of %d:_\n\n",
	},
	MsgCheckClientMissingItemFmt: {
		LocaleRU: "• `%s`\n  `%s`\n",
		LocaleEN: "• `%s`\n  `%s`\n",
	},
	MsgCheckClientDurationFooterFmt: {
		LocaleRU: "\n⏱ Время выполнения: *%s*",
		LocaleEN: "\n⏱ Duration: *%s*",
	},
	MsgMainMenuClientConnectErrorFmt: {
		LocaleRU: "❌ *%s* - ошибка подключения\n\n",
		LocaleEN: "❌ *%s* - connection error\n\n",
	},
	MsgMainMenuClientTransferInfoErrorFmt: {
		LocaleRU: "⚠️ *%s* - ошибка получения данных\n\n",
		LocaleEN: "⚠️ *%s* - failed to get data\n\n",
	},
	MsgMainMenuClientLinePrefixFmt: {
		LocaleRU: "🔹 *%s* ",
		LocaleEN: "🔹 *%s* ",
	},
	MsgMainMenuDownloadFmt: {
		LocaleRU: "⬇️ %s",
		LocaleEN: "⬇️ %s",
	},
	MsgMainMenuUploadFmt: {
		LocaleRU: "⬆️ %s",
		LocaleEN: "⬆️ %s",
	},
	MsgMainMenuSpeedLimitBracketFmt: {
		LocaleRU: " \\[%s]",
		LocaleEN: " \\[%s]",
	},
	MsgMainMenuNoClientsText: {
		LocaleRU: "🏠 *Главное меню*\n\nКлиенты не найдены.\nДобавьте клиента: *Настройки* → *Клиенты*.",
		LocaleEN: "🏠 *Main menu*\n\nNo clients found.\nAdd a client: *Settings* → *Clients*.",
	},
	MsgSpeedZero: {
		LocaleRU: "0 B/s",
		LocaleEN: "0 B/s",
	},
	MsgSpeedBpsFmt: {
		LocaleRU: "%d B/s",
		LocaleEN: "%d B/s",
	},
	MsgSpeedKBpsFmt: {
		LocaleRU: "%.1f KB/s",
		LocaleEN: "%.1f KB/s",
	},
	MsgSpeedMBpsFmt: {
		LocaleRU: "%.1f МБ/с",
		LocaleEN: "%.1f MB/s",
	},
	MsgSpeedLimitMBpsFmt: {
		LocaleRU: "%.2f МБ/с",
		LocaleEN: "%.2f MB/s",
	},
	MsgClientListItemFmt: {
		LocaleRU: "%s *%s*\n",
		LocaleEN: "%s *%s*\n",
	},
	MsgClientHostPortFmt: {
		LocaleRU: "   `%s:%d`\n\n",
		LocaleEN: "   `%s:%d`\n\n",
	},
	MsgClientButtonLabelFmt: {
		LocaleRU: "%s %s",
		LocaleEN: "%s %s",
	},
	MsgSettingsMenuText: {
		LocaleRU: "⚙️ *Настройки*\n\nВыберите раздел настроек:",
		LocaleEN: "⚙️ *Settings*\n\nChoose a settings section:",
	},
	MsgNotificationsStatusOn: {
		LocaleRU: "ВКЛ",
		LocaleEN: "ON",
	},
	MsgNotificationsStatusOff: {
		LocaleRU: "ВЫКЛ",
		LocaleEN: "OFF",
	},
	MsgNotificationsToggleButtonFmt: {
		LocaleRU: "🔔 Уведомления: %s",
		LocaleEN: "🔔 Notifications: %s",
	},
	MsgTimezoneMenuText: {
		LocaleRU: "🕒 *Часовой пояс*\n\nВыберите часовой пояс или отправьте свою локацию:",
		LocaleEN: "🕒 *Timezone*\n\nChoose a timezone or send your location:",
	},
	MsgVariablesMenuText: {
		LocaleRU: "🔧 *Переменные*",
		LocaleEN: "🔧 *Variables*",
	},
	MsgKeyValueIntFmt: {
		LocaleRU: "%s: %d",
		LocaleEN: "%s: %d",
	},
	MsgKeyValueStringFmt: {
		LocaleRU: "%s: %s",
		LocaleEN: "%s: %s",
	},
	MsgEditRecommendedTorrentsTextFmt: {
		LocaleRU: "🔧 *Редактирование переменной*\n\nРекомендуемое количество торрентов на странице выбора мониторинга: *%d*\n\nВыберите новое значение:",
		LocaleEN: "🔧 *Edit variable*\n\nRecommended torrents count on monitoring selection page: *%d*\n\nChoose a new value:",
	},
	MsgSearchStartPromptText: {
		LocaleRU: "🔎 *Поиск торрента*\n\nВведите хеш или название торрента (частичное или полное):",
		LocaleEN: "🔎 *Torrent search*\n\nEnter hash or torrent name (partial or full):",
	},
	MsgSearchClientsListError: {
		LocaleRU: "❌ Ошибка при получении списка клиентов",
		LocaleEN: "❌ Failed to load clients list",
	},
	MsgSearchNoClientsText: {
		LocaleRU: "🔎 *Поиск торрента*\n\nКлиенты не найдены. Добавьте клиента для поиска.",
		LocaleEN: "🔎 *Torrent search*\n\nNo clients found. Add a client to search.",
	},
	MsgSearchResultsStaleText: {
		LocaleRU: "Результаты поиска устарели. Выполните поиск заново.",
		LocaleEN: "Search results are stale. Start a new search.",
	},
	MsgSearchResultsButtonItemFmt: {
		LocaleRU: "%d. %s",
		LocaleEN: "%d. %s",
	},
	MsgQuickActionsMenuText: {
		LocaleRU: "⚡ *Быстрые действия*\n\nВыберите действие:",
		LocaleEN: "⚡ *Quick actions*\n\nChoose an action:",
	},
	MsgQuickActionsMenuNoClientsText: {
		LocaleRU: "⚡ *Быстрые действия*\n\nКлиенты не найдены. Добавьте клиента для использования быстрых действий.",
		LocaleEN: "⚡ *Quick actions*\n\nNo clients found. Add a client to use quick actions.",
	},
	MsgQuickActionsPauseMenuText: {
		LocaleRU: "⏸ *Остановка раздач*\n\nВыберите какие раздачи остановить:",
		LocaleEN: "⏸ *Pause torrents*\n\nChoose which torrents to pause:",
	},
	MsgQuickActionsResumeMenuText: {
		LocaleRU: "▶ *Возобновление раздач*\n\nВыберите какие раздачи запустить:",
		LocaleEN: "▶ *Resume torrents*\n\nChoose which torrents to resume:",
	},
	MsgSpeedLimitMenuText: {
		LocaleRU: "🚦 *Ограничение скорости*\n\nВыберите скорость:",
		LocaleEN: "🚦 *Speed limit*\n\nChoose a speed:",
	},
	MsgSpeedLimitMenuNoClientsText: {
		LocaleRU: "🚦 *Ограничение скорости*\n\nКлиенты не найдены. Добавьте клиента для использования ограничения скорости.",
		LocaleEN: "🚦 *Speed limit*\n\nNo clients found. Add a client to use speed limits.",
	},
	MsgSpeedLimitCustomPromptText: {
		LocaleRU: "🚦 *Ограничение скорости*\n\nВведите скорость в МБ/с (например: 2.5):",
		LocaleEN: "🚦 *Speed limit*\n\nEnter speed in MB/s (e.g. 2.5):",
	},
	MsgSpeedLimitAppliedHeaderFmt: {
		LocaleRU: "🚦 *Ограничение скорости до %.2f МБ/с*\n\n",
		LocaleEN: "🚦 *Speed limit to %.2f MB/s*\n\n",
	},
	MsgSpeedLimitsRemovedHeader: {
		LocaleRU: "🚫 *Снятие ограничений скорости*\n\n",
		LocaleEN: "🚫 *Removing speed limits*\n\n",
	},
	MsgPauseAllHeaderText: {
		LocaleRU: "⏸ *Остановка всех раздач*\n\n",
		LocaleEN: "⏸ *Pause all torrents*\n\n",
	},
	MsgResumeAllHeaderText: {
		LocaleRU: "▶ *Запуск всех раздач*\n\n",
		LocaleEN: "▶ *Resume all torrents*\n\n",
	},
	MsgPauseAllClientSuccessFmt: {
		LocaleRU: "✅ *%s* - остановлено\n",
		LocaleEN: "✅ *%s* - paused\n",
	},
	MsgResumeAllClientSuccessFmt: {
		LocaleRU: "✅ *%s* - запущено\n",
		LocaleEN: "✅ *%s* - resumed\n",
	},
	MsgPauseRutrackerHeaderText: {
		LocaleRU: "⏸ *Остановка rutracker раздач*\n\n",
		LocaleEN: "⏸ *Pause rutracker torrents*\n\n",
	},
	MsgResumeRutrackerHeaderText: {
		LocaleRU: "▶ *Запуск rutracker раздач*\n\n",
		LocaleEN: "▶ *Resume rutracker torrents*\n\n",
	},
	MsgPauseRutrackerClientSuccessFmt: {
		LocaleRU: "✅ *%s* - остановлено (%d)\n",
		LocaleEN: "✅ *%s* - paused (%d)\n",
	},
	MsgResumeRutrackerClientSuccessFmt: {
		LocaleRU: "✅ *%s* - запущено (%d)\n",
		LocaleEN: "✅ *%s* - resumed (%d)\n",
	},
	MsgPauseNonRutrackerHeaderText: {
		LocaleRU: "⏸ *Остановка НЕ rutracker раздач*\n\n",
		LocaleEN: "⏸ *Pause non-rutracker torrents*\n\n",
	},
	MsgResumeNonRutrackerHeaderText: {
		LocaleRU: "▶ *Запуск НЕ rutracker раздач*\n\n",
		LocaleEN: "▶ *Resume non-rutracker torrents*\n\n",
	},
	MsgPauseNonRutrackerClientSuccessFmt: {
		LocaleRU: "✅ *%s* - остановлено (%d)\n",
		LocaleEN: "✅ *%s* - paused (%d)\n",
	},
	MsgResumeNonRutrackerClientSuccessFmt: {
		LocaleRU: "✅ *%s* - запущено (%d)\n",
		LocaleEN: "✅ *%s* - resumed (%d)\n",
	},
	MsgResultErrorsHeaderFmt: {
		LocaleRU: "\n❌ Ошибки (%d):\n",
		LocaleEN: "\n❌ Errors (%d):\n",
	},
	MsgResultErrorsItemFmt: {
		LocaleRU: "  • %s\n",
		LocaleEN: "  • %s\n",
	},
	MsgResultTotalsFmt: {
		LocaleRU: "\nВсего обработано: %d успешно, %d с ошибками",
		LocaleEN: "\nProcessed total: %d success, %d failed",
	},
	MsgBotCommandMenuDescription: {
		LocaleRU: "Главное меню",
		LocaleEN: "Main menu",
	},
	MsgBotCommandCheckDescription: {
		LocaleRU: "Проверить статус",
		LocaleEN: "Check status",
	},
	MsgBotCommandClientsDescription: {
		LocaleRU: "Управление клиентами qBittorrent",
		LocaleEN: "Manage qBittorrent clients",
	},
	MsgTorrentMonitorSelectText: {
		LocaleRU: "📊 *Мониторинг торрента*\n\nВыберите торрент или введите хеш вручную:",
		LocaleEN: "📊 *Torrent monitoring*\n\nChoose a torrent or enter hash manually:",
	},
	MsgTorrentMonitorManualHashPrompt: {
		LocaleRU: "📊 *Мониторинг торрента*\n\nВведите хеш торрента для мониторинга:",
		LocaleEN: "📊 *Torrent monitoring*\n\nEnter torrent hash to monitor:",
	},
	MsgTorrentMonitorItemButtonFmt: {
		LocaleRU: "📁 %s",
		LocaleEN: "📁 %s",
	},
	MsgSavePathSelectionHeaderText: {
		LocaleRU: "📁 *Выберите путь сохранения*\n\n",
		LocaleEN: "📁 *Choose save path*\n\n",
	},
	MsgSavePathSelectionTorrentFmt: {
		LocaleRU: "Торрент: `%s`\n\n",
		LocaleEN: "Torrent: `%s`\n\n",
	},
	MsgSavePathSelectionRecommendedBlockFmt: {
		LocaleRU: "⭐ *Рекомендуется* (используется для этого торрента):\n`%s`\n\n",
		LocaleEN: "⭐ *Recommended* (used for this torrent):\n`%s`\n\n",
	},
	MsgSavePathSelectionRecommendedButtonFmt: {
		LocaleRU: "⭐ Рекомендуется: %s",
		LocaleEN: "⭐ Recommended: %s",
	},
	MsgSavePathSelectionDefaultBlockFmt: {
		LocaleRU: "По умолчанию: `%s`\n\n",
		LocaleEN: "Default: `%s`\n\n",
	},
	MsgSavePathSelectionDefaultButtonFmt: {
		LocaleRU: "📂 По умолчанию (%s)",
		LocaleEN: "📂 Default (%s)",
	},
	MsgSavePathSelectionExistingPathsHeaderText: {
		LocaleRU: "Существующие пути:\n",
		LocaleEN: "Existing paths:\n",
	},
	MsgSavePathSelectionPathButtonFmt: {
		LocaleRU: "📂 %s",
		LocaleEN: "📂 %s",
	},
	MsgCustomSavePathPromptText: {
		LocaleRU: "✏️ *Ввод пути сохранения*\n\nВведите путь для сохранения торрента:",
		LocaleEN: "✏️ *Save path input*\n\nEnter save path for the torrent:",
	},
	MsgSkipHashCheckQuestionFmt: {
		LocaleRU: "⚙️ *Настройки добавления торрента*\n\nПуть сохранения: `%s`\n\n❓ Пропустить проверку хеша при добавлении?\n\n_Проверка хеша может занять время, но гарантирует целостность данных._",
		LocaleEN: "⚙️ *Torrent add settings*\n\nSave path: `%s`\n\n❓ Skip hash check while adding?\n\n_Hash check may take time, but it ensures data integrity._",
	},
	MsgDeleteExistingTorrentQuestionFmt: {
		LocaleRU: "✅ \n\n⚠️ Найден существующий торрент с таким же названием:\n`%s`\n\n❓ Удалить старый торрент?",
		LocaleEN: "✅ \n\n⚠️ Found an existing torrent with the same name:\n`%s`\n\n❓ Delete the old torrent?",
	},
	MsgDeleteFilesQuestionText: {
		LocaleRU: "🗑️ *Удаление торрента*\n\n❓ Удалить файлы вместе с торрентом?\n\n_Если выбрать \"Да\", файлы будут удалены с диска._",
		LocaleEN: "🗑️ *Delete torrent*\n\n❓ Delete files along with the torrent?\n\n_If you choose \"Yes\", files will be deleted from disk._",
	},
	MsgTorrentDeletedFilesSuffix: {
		LocaleRU: " и файлы",
		LocaleEN: " and files",
	},
	MsgTorrentDeletedSuccessFmt: {
		LocaleRU: "✅ Торрент успешно удален%s из клиента *%s*",
		LocaleEN: "✅ Torrent deleted%s from client *%s*",
	},
	MsgAddTorrentSendFilePromptFmt: {
		LocaleRU: "📥 *Добавление торрент файла*\n\nКлиент: *%s*\n\n📎 Отправьте торрент файл (.torrent):",
		LocaleEN: "📥 *Add torrent file*\n\nClient: *%s*\n\n📎 Send the torrent file (.torrent):",
	},
	MsgTorrentProgressHeaderText: {
		LocaleRU: "📊 *Прогресс торрента*\n\n",
		LocaleEN: "📊 *Torrent progress*\n\n",
	},
	MsgTorrentProgressNameFmt: {
		LocaleRU: " `%s`\n\n",
		LocaleEN: " `%s`\n\n",
	},
	MsgTorrentProgressPathFmt: {
		LocaleRU: "📁 `%s`\n\n",
		LocaleEN: "📁 `%s`\n\n",
	},
	MsgTorrentProgressStatusFmt: {
		LocaleRU: "Статус: %s\n",
		LocaleEN: "Status: %s\n",
	},
	MsgTorrentProgressPercentFmt: {
		LocaleRU: "Прогресс: *%.1f%%*\n\n",
		LocaleEN: "Progress: *%.1f%%*\n\n",
	},
	MsgTorrentProgressDownloadFmt: {
		LocaleRU: "⬇️ Загрузка: %s\n",
		LocaleEN: "⬇️ Download: %s\n",
	},
	MsgTorrentProgressUploadFmt: {
		LocaleRU: "⬆️ Отдача: %s\n",
		LocaleEN: "⬆️ Upload: %s\n",
	},
	MsgTorrentProgressUploadedFmt: {
		LocaleRU: "📤 Всего отдано: %s\n",
		LocaleEN: "📤 Uploaded total: %s\n",
	},
	MsgTorrentProgressSeedsPeersFmt: {
		LocaleRU: "👥 Сиды: %d | Пиры: %d\n\n",
		LocaleEN: "👥 Seeds: %d | Peers: %d\n\n",
	},
	MsgTorrentProgressSizeFmt: {
		LocaleRU: "📦 Размер: %s / %s",
		LocaleEN: "📦 Size: %s / %s",
	},
	MsgTorrentProgressSpeedSuffixPerSec: {
		LocaleRU: "/s",
		LocaleEN: "/s",
	},
	MsgTorrentStatusCompleted: {
		LocaleRU: "✅ Завершен",
		LocaleEN: "✅ Completed",
	},
	MsgTorrentStatusPaused: {
		LocaleRU: "⏸ Приостановлен",
		LocaleEN: "⏸ Paused",
	},
	MsgTorrentStatusDownloading: {
		LocaleRU: "▶ Загружается",
		LocaleEN: "▶ Downloading",
	},
	MsgTorrentStatusError: {
		LocaleRU: "⚠️ Ошибка",
		LocaleEN: "⚠️ Error",
	},
	MsgTorrentStatusMissingFiles: {
		LocaleRU: "⚠️ Отсутствуют файлы",
		LocaleEN: "⚠️ Missing files",
	},
	MsgTorrentStatusOtherFmt: {
		LocaleRU: "ℹ️ %s",
		LocaleEN: "ℹ️ %s",
	},
	MsgTorrentMonitorClientProcessingFmt: {
		LocaleRU: "✅\n\nКлиент: *%s*\n\n⏳ Обработка...",
		LocaleEN: "✅\n\nClient: *%s*\n\n⏳ Processing...",
	},
	MsgTorrentMonitorClientTorrentProcessingFmt: {
		LocaleRU: "✅\n\nКлиент: *%s*\n\n⏳ Торрент обрабатывается qBittorrent...\n\n_Обработка..._",
		LocaleEN: "✅\n\nClient: *%s*\n\n⏳ Torrent is being processed by qBittorrent...\n\n_Processing..._",
	},
	MsgTorrentMonitorOpenTorrentButtonText: {
		LocaleRU: "🔗 Открыть раздачу",
		LocaleEN: "🔗 Open torrent",
	},
	MsgTorrentActivityDownloading: {
		LocaleRU: "⬇️ Загрузка",
		LocaleEN: "⬇️ Downloading",
	},
	MsgTorrentActivityUploading: {
		LocaleRU: "⬆️ Раздача",
		LocaleEN: "⬆️ Uploading",
	},
	MsgTorrentActivityUploadStalled: {
		LocaleRU: "⚠️ Раздача (застой)",
		LocaleEN: "⚠️ Uploading (stalled)",
	},
	MsgTorrentActivityDownloadStalled: {
		LocaleRU: "⚠️ Загрузка (застой)",
		LocaleEN: "⚠️ Downloading (stalled)",
	},
	MsgTorrentActivityChecking: {
		LocaleRU: "🔍 Проверка",
		LocaleEN: "🔍 Checking",
	},
	MsgTorrentActivityQueued: {
		LocaleRU: "⏳ В очереди",
		LocaleEN: "⏳ Queued",
	},
	MsgTorrentActivityPaused: {
		LocaleRU: "⏸ Остановлен",
		LocaleEN: "⏸ Paused",
	},
	MsgTorrentActivityFetchingMetadata: {
		LocaleRU: "⬇️ Получение метаданных",
		LocaleEN: "⬇️ Fetching metadata",
	},
	MsgTorrentActivityError: {
		LocaleRU: "❌ Ошибка",
		LocaleEN: "❌ Error",
	},
	MsgTorrentActivityMissingFiles: {
		LocaleRU: "⚠️ Отсутствуют файлы",
		LocaleEN: "⚠️ Missing files",
	},
	MsgTorrentActivityOtherFmt: {
		LocaleRU: "ℹ️ %s",
		LocaleEN: "ℹ️ %s",
	},
}

func Msg(id MsgID) string {
	return MsgForLocale(id, DefaultLocale)
}

func MsgForLocale(id MsgID, locale Locale) string {
	res, ok := msgResources[id]
	if !ok {
		return string(id)
	}
	if txt, hasText := res[locale]; hasText && txt != "" {
		return txt
	}
	if txt, hasText := res[DefaultLocale]; hasText && txt != "" {
		return txt
	}

	return string(id)
}

func Msgs(id MsgID, args ...any) string {
	return fmt.Sprintf(Msg(id), args...)
}

func Text(id ButtonID) string {
	return TextForLocale(id, DefaultLocale)
}

func TextForLocale(id ButtonID, locale Locale) string {
	res, ok := buttonResources[id]
	if !ok {
		return string(id)
	}
	if txt, hasText := res.text[locale]; hasText && txt != "" {
		return txt
	}
	if txt, hasText := res.text[DefaultLocale]; hasText && txt != "" {
		return txt
	}

	return string(id)
}

func Button(id ButtonID) tgbotapi.InlineKeyboardButton {
	return ButtonForLocale(id, DefaultLocale)
}

func ButtonForLocale(id ButtonID, locale Locale) tgbotapi.InlineKeyboardButton {
	res, ok := buttonResources[id]
	if !ok {
		return tgbotapi.NewInlineKeyboardButtonData(string(id), string(id))
	}
	data := res.data
	if data == "" {
		data = string(id)
	}

	return tgbotapi.NewInlineKeyboardButtonData(TextForLocale(id, locale), data)
}

func ButtonWithData(id ButtonID, callbackData string) tgbotapi.InlineKeyboardButton {
	return ButtonWithDataForLocale(id, callbackData, DefaultLocale)
}

func ButtonWithDataForLocale(id ButtonID, callbackData string, locale Locale) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(TextForLocale(id, locale), callbackData)
}

func Data(text, callbackData string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(text, callbackData)
}
