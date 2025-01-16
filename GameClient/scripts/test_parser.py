import unittest

from src.backend.parser import parse_message
from src.shared.constants import NetworkMessage, Param


class TestParseMessage(unittest.TestCase):
    def test_parse_message_valid(self):
        input_str = "KIVUPS01" + "2025-01-05 12:00:00.000000" + "{test_player}" + "{\"paramName\":\"paramValue\"}\n"
        expected_message = NetworkMessage(
            signature="KIVUPS",
            command_id=1,
            timestamp="2025-01-05 12:00:00.000000",
            player_nickname="test_player",
            parameters=[Param(name="paramName", value="paramValue")]
        )
        result = parse_message(input_str)
        self.assertEqual(result, expected_message)

    def test_parse_message_invalid_signature(self):
        input_str = "INVALID01" + "2025-01-05 12:00:00.000000" + "{test_player}" + "{\"paramName\":\"paramValue\"}\n"
        with self.assertRaises(ValueError):
            parse_message(input_str)

    def test_parse_message_invalid_command_id(self):
        input_str = "KIVUPS99" + "2025-01-05 12:00:00.000000" + "{test_player}" + "{\"paramName\":\"paramValue\"}\n"
        with self.assertRaises(ValueError):
            parse_message(input_str)

    def test_parse_message_invalid_format(self):
        input_str = "KIVUPS01" + "2025-01-05 12:00:00.000000" + "{test_player}" + "{\"paramName\":\"paramValue\"}"
        with self.assertRaises(ValueError):
            parse_message(input_str)

    # test no param Value
    def test_parse_message_no_param_value(self):
        input_str = "KIVUPS01" + "2025-01-05 12:00:00.000000" + "{test_player}" + "{\"paramName\":\"\"}\n"
        expected_message = NetworkMessage(
            signature="KIVUPS",
            command_id=1,
            timestamp="2025-01-05 12:00:00.000000",
            player_nickname="test_player",
            parameters=[Param(name="paramName", value="")]
        )
        result = parse_message(input_str)
        self.assertEqual(result, expected_message)

    # test param list gamelist
    def test_parse_message_param_list(self):
        input_str = "KIVUPS01" + "2025-01-05 12:00:00.000000" + "{test_player}" + "\"{\"gameList\":[{\"gameName\":\"Game1\",\"maxPlayers\":\"4\",\"connectedPlayers\":\"2\"};{\"gameName\":\"Game2\",\"maxPlayers\":\"4\",\"connectedPlayers\":\"2\"}]\"}\n"
        expected_message = NetworkMessage(
            signature="KIVUPS",
            command_id=1,
            timestamp="2025-01-05 12:00:00.000000",
            player_nickname="test_player",
            parameters=[Param(name="gameList", value="[{\"gameName\":\"Game1\",\"maxPlayers\":\"4\",\"connectedPlayers\":\"2\"},{\"gameName\":\"Game2\",\"maxPlayers\":\"4\",\"connectedPlayers\":\"2\"}")]
        )
        result = parse_message(input_str)
        self.assertEqual(result, expected_message)

if __name__ == "__main__":
    unittest.main()