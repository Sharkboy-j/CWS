package dialogstate

type State string

const (
	StateAddClientName      State = "add_client_name"
	StateAddClientHost      State = "add_client_host"
	StateAddClientPort      State = "add_client_port"
	StateAddClientUsername  State = "add_client_username"
	StateAddClientPassword  State = "add_client_password"
	StateAddClientSSL       State = "add_client_ssl"
	StateEditClientName     State = "edit_client_name"
	StateEditClientHost     State = "edit_client_host"
	StateEditClientPort     State = "edit_client_port"
	StateEditClientUsername State = "edit_client_username"
	StateEditClientPassword State = "edit_client_password"
	StateEditClientSSL      State = "edit_client_ssl"
	StateAddTorrentCustom   State = "add_torrent_custom_path"
	StateAddTorrentWait     State = "add_torrent_wait_file"
	StateMonitorTorrent     State = "monitor_torrent_hash"
	StateSearchTorrent      State = "search_torrent_query"
	StateCustomSpeedLimit   State = "custom_speed_limit"
	StateEditRecommended    State = "edit_recommended_torrents_input"
)
