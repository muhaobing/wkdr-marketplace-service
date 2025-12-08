Dim WshShell
Set WshShell = WScript.CreateObject("WScript.Shell")
WshShell.SendKeys "ssh root@182.92.168.209 -p 22"
WshShell.SendKeys "{ENTER}"
WScript.Sleep 1000
WshShell.SendKeys "ECS_powermind"
WshShell.SendKeys "{ENTER}"
